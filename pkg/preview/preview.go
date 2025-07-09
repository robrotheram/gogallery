package preview

import (
	"compress/gzip"
	"gogallery/pkg/config"
	"gogallery/pkg/embeds"
	"gogallery/pkg/pipeline"
	templateengine "gogallery/pkg/templateEngine"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

const assetPrefix = "/assets/"

/**
dynmically gnerate page for previewing site
**/

// cacheMiddleware sets cache headers for static assets and API responses.
func cacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Cache for 1 hour for assets, otherwise no-cache
		if strings.HasPrefix(r.URL.Path, assetPrefix) || strings.HasPrefix(r.URL.Path, "/img/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000")
		}
		next.ServeHTTP(w, r)
	})
}

// compressionMiddleware compresses HTTP responses using gzip if the client supports it.
func compressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		next.ServeHTTP(gzw, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

var cfg = config.Config

func (api *Server) Setup() {

	api.Use(mux.CORSMethodMiddleware(api.Router))
	api.Use(cacheMiddleware)       // Add cache middleware
	api.Use(compressionMiddleware) // Add compression middleware

	api.HandleFunc("/", api.PreviewPageHandler).Methods("GET")
	api.HandleFunc("/img/{id}", api.ImgHandler)
	api.HandleFunc("/img/{id}/{size}.{ext}", api.ImgHandler)
	api.HandleFunc("/manifest.json", api.PreviewManifest).Methods("GET")
	api.HandleFunc("/albums", api.PreviewAlbumsHandler).Methods("GET")
	api.HandleFunc("/photo/{id}", api.PreviewPictureHandler).Methods("GET")
	api.HandleFunc("/album/{id}", api.PreviewCollectionHandler).Methods("GET")

	api.PathPrefix(assetPrefix).Handler(api.assestHandler())
}
func (api *Server) assestHandler() http.Handler {
	if embeds.DoesThmeExist(cfg.Gallery.Theme) {
		return templateengine.Templates.AsseetServer(cfg.Gallery.Theme, assetPrefix)
	}
	asestPath := config.Config.Gallery.Theme + assetPrefix
	return http.StripPrefix(assetPrefix, http.FileServer(http.Dir(asestPath)))
}

func (api *Server) ImgHandler(w http.ResponseWriter, r *http.Request) {
	size := r.URL.Query().Get("size")
	vars := mux.Vars(r)
	id := vars["id"]
	if len(size) == 0 {
		size = vars["size"]
	}
	pic, _ := api.Pictures.FindById(id)
	//Is image in cache
	if file, err := api.ImageCache.Get(pic.Id, config.WebP, size); err == nil {
		io.Copy(w, file)
		return
	}

	src, err := pic.Load()
	if err != nil {
		return
	}
	cache, _ := api.ImageCache.Writer(pic.Id, config.WebP, size)
	writer := io.MultiWriter(w, cache)
	if size, ok := templateengine.ImageSizes[size]; ok {
		pipeline.ProcessImage(src, size.ImgWidth, config.WebP, writer)
		return
	}
}

func (api *Server) PreviewPageHandler(w http.ResponseWriter, r *http.Request) {
	templateengine.Templates.Load(cfg.Gallery.Theme)
	builder := pipeline.NewRenderPipeline(&cfg.Gallery, api.DataStore)

	w.WriteHeader(200)
	builder.BuildIndex(w)
}

func (api *Server) PreviewAlbumsHandler(w http.ResponseWriter, r *http.Request) {
	templateengine.Templates.Load(cfg.Gallery.Theme)
	builder := pipeline.NewRenderPipeline(&cfg.Gallery, api.DataStore)
	w.WriteHeader(200)
	builder.BuildAlbums(w)
}

func (api *Server) PreviewPictureHandler(w http.ResponseWriter, r *http.Request) {
	templateengine.Templates.Load(cfg.Gallery.Theme)
	builder := pipeline.NewRenderPipeline(&cfg.Gallery, api.DataStore)
	photoID := mux.Vars(r)["id"]
	pic, _ := api.Pictures.FindById(photoID)
	w.WriteHeader(200)
	builder.BuildPhoto(pic, w)
}

func (api *Server) PreviewCollectionHandler(w http.ResponseWriter, r *http.Request) {
	templateengine.Templates.Load(cfg.Gallery.Theme)
	builder := pipeline.NewRenderPipeline(&cfg.Gallery, api.DataStore)
	photoID := mux.Vars(r)["id"]
	w.WriteHeader(200)
	builder.BuildAlbum(photoID, w)
}

func (api *Server) PreviewManifest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	templateengine.ManifestWriter(w, &cfg.Gallery)
}
