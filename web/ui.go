package web

import (
	"encoding/json"
	"log"
	"net/http"

	_ "net/http/pprof"

	"github.com/gorilla/mux"
	galleryConfig "github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
)

var ViewCount = 0
var config *galleryConfig.Configuration

func Serve(conf *galleryConfig.Configuration) {
	config = conf
	r := mux.NewRouter()
	r.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	r.HandleFunc("/albums", renderAlbum)
	r.HandleFunc("/album/pic/{picture}", renderAlbumPicturePage)
	r.HandleFunc("/album/{name}", renderAlbumPage)
	r.HandleFunc("/album/{name}/{page}", renderAlbumPagination)
	r.HandleFunc("/pic/{picture}", renderPicturePage)
	r.HandleFunc("/img/{name}", loadImage)
	r.HandleFunc("/thumb/{name}", loadThumbnail)
	r.HandleFunc("/thumb/{name}/large", loadLargeThumbnail)

	r.HandleFunc("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(makeManifest())
	})
	r.HandleFunc("/sw.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, themePath()+"static/js/sw.js")
	})

	if config.About.Enable {
		r.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
			renderGalleryTemplate(w, "aboutPage", nil, datastore.Picture{}, 0)
		})
	}

	registerAdmin(r)

	r.HandleFunc("/", renderIndexPage)
	r.HandleFunc("/error", renderErrorPage)
	r.HandleFunc("/{name}", renderIndexPaginationPage)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", CacheControlWrapper(http.FileServer(http.Dir(themePath()+"static")))))

	log.Println("Starting server on port" + config.Server.Port)
	log.Fatal(http.ListenAndServe(config.Server.Port, (r)))
}
