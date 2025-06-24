package api

import (
	"compress/gzip"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/backend/pipeline"
	templateengine "github.com/robrotheram/gogallery/backend/templateEngine"
)

/**
dynmically gnerate page for previewing site
**/

// cacheMiddleware sets cache headers for static assets and API responses.
func cacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Cache for 1 hour for assets, otherwise no-cache
		if strings.HasPrefix(r.URL.Path, "/assets/") || strings.HasPrefix(r.URL.Path, "/img/") {
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

func (api *GoGalleryAPI) ServePreviewSite() {

	api.router.Use(mux.CORSMethodMiddleware(api.router))
	api.router.Use(cacheMiddleware)       // Add cache middleware
	api.router.Use(compressionMiddleware) // Add compression middleware

	api.router.HandleFunc("/", api.PreviewPageHandler).Methods("GET")
	api.router.HandleFunc("/manifest.json", api.PreviewManifest).Methods("GET")
	api.router.HandleFunc("/albums", api.PreviewAlbumsHandler).Methods("GET")
	api.router.HandleFunc("/photo/{id}", api.PreviewPictureHandler).Methods("GET")
	api.router.HandleFunc("/album/{id}", api.PreviewCollectionHandler).Methods("GET")

	assestPath := api.config.Gallery.Theme + "/assets/"
	api.router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir(assestPath))))

	log.Println("Starting server on port: http://" + api.config.Server.GetAddr())
	log.Fatal(http.ListenAndServe(api.config.Server.GetAddr(), api.router))
}

func (api *GoGalleryAPI) PreviewPageHandler(w http.ResponseWriter, r *http.Request) {
	templateengine.Templates.Load(api.config.Gallery.Theme)
	builder := pipeline.NewRenderPipeline(&api.config.Gallery, api.DataStore, api.monitor)

	w.WriteHeader(200)
	builder.BuildIndex(w)
}

func (api *GoGalleryAPI) PreviewAlbumsHandler(w http.ResponseWriter, r *http.Request) {
	templateengine.Templates.Load(api.config.Gallery.Theme)
	builder := pipeline.NewRenderPipeline(&api.config.Gallery, api.DataStore, api.monitor)
	w.WriteHeader(200)
	builder.BuildAlbums(w)
}

func (api *GoGalleryAPI) PreviewPictureHandler(w http.ResponseWriter, r *http.Request) {
	templateengine.Templates.Load(api.config.Gallery.Theme)
	builder := pipeline.NewRenderPipeline(&api.config.Gallery, api.DataStore, api.monitor)
	photoID := mux.Vars(r)["id"]
	pic, _ := api.Pictures.FindById(photoID)
	w.WriteHeader(200)
	builder.BuildPhoto(pic, w)
}

func (api *GoGalleryAPI) PreviewCollectionHandler(w http.ResponseWriter, r *http.Request) {
	templateengine.Templates.Load(api.config.Gallery.Theme)
	builder := pipeline.NewRenderPipeline(&api.config.Gallery, api.DataStore, api.monitor)
	photoID := mux.Vars(r)["id"]
	w.WriteHeader(200)
	builder.BuildAlbum(photoID, w)
}

func (api *GoGalleryAPI) PreviewManifest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	templateengine.ManifestWriter(w, &api.config.Gallery)
}
