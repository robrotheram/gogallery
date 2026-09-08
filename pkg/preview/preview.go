package preview

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"gogallery/pkg/embeds"
	"gogallery/pkg/pipeline"
	templateengine "gogallery/pkg/templateEngine"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

const assetPrefix = "/assets/"

// Dynamically generate pages for previewing the site.

// cacheMiddleware sets cache headers for static assets and API responses.
func cacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, assetPrefix):
			w.Header().Set("Cache-Control", "public, max-age=86400")
		case strings.HasPrefix(r.URL.Path, "/img/"):
			w.Header().Set("Cache-Control", "public, max-age=3600")
		default:
			w.Header().Set("Cache-Control", "no-cache")
		}
		next.ServeHTTP(w, r)
	})
}

// compressionMiddleware compresses HTTP responses using gzip if the client supports it.
func compressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Accept-Encoding")
		if strings.HasPrefix(r.URL.Path, "/img/") || r.Header.Get("Range") != "" ||
			!strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
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

func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		next.ServeHTTP(w, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (api *Server) Setup() {

	api.Use(mux.CORSMethodMiddleware(api.Router))
	api.Use(securityHeadersMiddleware)
	api.Use(cacheMiddleware)       // Add cache middleware
	api.Use(compressionMiddleware) // Add compression middleware

	api.HandleFunc("/", api.PreviewPageHandler).Methods("GET")
	api.HandleFunc("/img/{id}", api.ImgHandler).Methods(http.MethodGet, http.MethodHead)
	api.HandleFunc("/img/{id}/{size}.{ext}", api.ImgHandler).Methods(http.MethodGet, http.MethodHead)
	api.HandleFunc("/manifest.json", api.PreviewManifest).Methods("GET")
	api.HandleFunc("/albums", api.PreviewAlbumsHandler).Methods("GET")
	api.HandleFunc("/photo/{id}", api.PreviewPictureHandler).Methods("GET")
	api.HandleFunc("/album/{id}", api.PreviewCollectionHandler).Methods("GET")

	api.PathPrefix(assetPrefix).Handler(api.assetHandler())
}
func (api *Server) assetHandler() http.Handler {
	theme := config.NormalizeTheme(config.Config.Gallery.Theme)
	if embeds.DoesThemeExist(theme) {
		return templateengine.Templates.AssetServer(theme, assetPrefix)
	}
	asestPath := theme + assetPrefix
	return http.StripPrefix(assetPrefix, http.FileServer(http.Dir(asestPath)))
}

func (api *Server) ImgHandler(w http.ResponseWriter, r *http.Request) {
	size := r.URL.Query().Get("size")
	vars := mux.Vars(r)
	id := vars["id"]
	if len(size) == 0 {
		size = vars["size"]
	}
	if size == "" {
		size = "small"
	}
	if id == "" || len(id) > 128 {
		http.Error(w, "invalid image ID", http.StatusBadRequest)
		return
	}
	if size != "original" {
		if _, ok := templateengine.ImageSizes[size]; !ok || (vars["ext"] != "" && vars["ext"] != "webp") {
			http.Error(w, "invalid image size or format", http.StatusBadRequest)
			return
		}
	}
	pic, err := api.Pictures.FindByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if !datastore.IsPicturePublishable(pic) {
		http.NotFound(w, r)
		return
	}
	if size == "original" {
		api.serveOriginalImage(w, r, pic)
		return
	}
	imageSize := templateengine.ImageSizes[size]
	w.Header().Set("Content-Type", "image/webp")
	//Is image in cache
	if file, err := api.ImageCache.Get(pic.Id, config.WebP, size); err == nil {
		defer file.Close()
		stat, statErr := file.Stat()
		if statErr != nil {
			writePreviewError(w, statErr)
			return
		}
		http.ServeContent(w, r, filepath.Base(file.Name()), stat.ModTime(), file)
		return
	}

	src, err := pic.Load()
	if err != nil {
		writePreviewError(w, err)
		return
	}
	var encoded bytes.Buffer
	if err := pipeline.ProcessImage(src, imageSize.ImgWidth, config.WebP, &encoded); err != nil {
		writePreviewError(w, err)
		return
	}
	if cache, cacheErr := api.ImageCache.Writer(pic.Id, config.WebP, size); cacheErr != nil {
		log.Printf("Could not cache preview image %s: %v", pic.Id, cacheErr)
	} else {
		if _, cacheErr = cache.Write(encoded.Bytes()); cacheErr == nil {
			cacheErr = cache.Close()
		} else {
			_ = cache.Abort()
		}
		if cacheErr != nil {
			log.Printf("Could not cache preview image %s: %v", pic.Id, cacheErr)
		}
	}
	w.Header().Set("Content-Length", fmt.Sprint(encoded.Len()))
	_, _ = w.Write(encoded.Bytes())
}

func (api *Server) serveOriginalImage(w http.ResponseWriter, r *http.Request, pic datastore.Picture) {
	if !config.Config.Gallery.UseOriginal || !strings.EqualFold(varsExt(r), strings.TrimPrefix(pic.Ext, ".")) {
		http.NotFound(w, r)
		return
	}
	file, err := os.Open(pic.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		writePreviewError(w, err)
		return
	}
	contentType := mime.TypeByExtension(pic.Ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	http.ServeContent(w, r, filepath.Base(pic.Path), stat.ModTime(), file)
}

func varsExt(r *http.Request) string {
	return mux.Vars(r)["ext"]
}

func (api *Server) PreviewPageHandler(w http.ResponseWriter, r *http.Request) {
	if !loadPreviewTemplates(w) {
		return
	}
	builder := pipeline.NewRenderPipeline(&config.Config.Gallery, api.DataStore)

	if err := builder.BuildIndex(w); err != nil {
		writePreviewError(w, err)
	}
}

func (api *Server) PreviewAlbumsHandler(w http.ResponseWriter, r *http.Request) {
	if !loadPreviewTemplates(w) {
		return
	}
	builder := pipeline.NewRenderPipeline(&config.Config.Gallery, api.DataStore)
	if err := builder.BuildAlbums(w); err != nil {
		writePreviewError(w, err)
	}
}

func (api *Server) PreviewPictureHandler(w http.ResponseWriter, r *http.Request) {
	if !loadPreviewTemplates(w) {
		return
	}
	builder := pipeline.NewRenderPipeline(&config.Config.Gallery, api.DataStore)
	photoID := mux.Vars(r)["id"]
	pic, err := api.Pictures.FindByID(photoID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if !datastore.IsPicturePublishable(pic) {
		http.NotFound(w, r)
		return
	}
	if err := builder.BuildPhoto(pic, w); err != nil {
		writePreviewError(w, err)
	}
}

func (api *Server) PreviewCollectionHandler(w http.ResponseWriter, r *http.Request) {
	if !loadPreviewTemplates(w) {
		return
	}
	builder := pipeline.NewRenderPipeline(&config.Config.Gallery, api.DataStore)
	photoID := mux.Vars(r)["id"]
	album, err := api.Albums.FindById(photoID)
	if err != nil || datastore.IsAlbumInBlacklist(album.Name) {
		http.NotFound(w, r)
		return
	}
	if err := builder.BuildAlbum(photoID, w); err != nil {
		writePreviewError(w, err)
	}
}

func (api *Server) PreviewManifest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := templateengine.ManifestWriter(w, &config.Config.Gallery); err != nil {
		writePreviewError(w, err)
	}
}

func loadPreviewTemplates(w http.ResponseWriter) bool {
	theme := config.NormalizeTheme(config.Config.Gallery.Theme)
	if err := templateengine.Templates.Load(theme); err != nil {
		writePreviewError(w, fmt.Errorf("load theme %q: %w", theme, err))
		return false
	}
	return true
}

func writePreviewError(w http.ResponseWriter, err error) {
	log.Printf("Preview rendering failed: %v", err)
	http.Error(w, "Unable to render gallery preview. Check the configured theme.", http.StatusInternalServerError)
}
