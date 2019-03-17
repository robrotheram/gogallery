package web

import (
	"encoding/json"
	"github.com/gorilla/mux"
	galleryConfig "github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
	"log"
	"net/http"
)

var ViewCount = 0
var config *galleryConfig.Configuration

func Serve(conf *galleryConfig.Configuration) {
	config = conf
	r := mux.NewRouter()
	r.HandleFunc("/albums", renderAlbum)
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
		http.ServeFile(w, r, "web/"+themePath()+"static/js/sw.js")
	})

	if config.About.Enable {
		r.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
			renderGalleryTemplate(w, "aboutPage", nil, datastore.Picture{}, 0)
		})
	}

	registerAdmin(r)
	r.HandleFunc("/", renderIndexPage)
	r.HandleFunc("/{name}", renderIndexPaginationPage)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", CacheControlWrapper(http.FileServer(http.Dir(themePath()+"static")))))

	log.Println("Starting server on port" + config.Server.Port)
	log.Fatal(http.ListenAndServe(config.Server.Port, r))
}
