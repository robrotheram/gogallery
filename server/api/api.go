package api

import (
	"io/ioutil"

	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/auth"
	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
)

type Stats struct {
	Photos int
	Albums int
	Rubish int
}

func MakeStats() Stats {
	s := Stats{0, 0, 0}
	var pics []datastore.Picture
	var albms []datastore.Album
	datastore.Cache.DB.All(&pics)
	datastore.Cache.DB.All(&albms)
	s.Photos = len(pics)
	s.Albums = len(albms)
	files, _ := ioutil.ReadDir(Config.Gallery.Basepath + "/rubish")
	s.Rubish = len(files)
	return s
}

var Config *config.Configuration

func InitApiRoutes(r *mux.Router, config *config.Configuration) *mux.Router {
	Config = config

	r.Handle("/photos", auth.AuthMiddleware(getAllPhotosHandler))
	r.Handle("/photo/{id}", auth.AuthMiddleware(getPhotoHandler)).Methods("GET")
	r.Handle("/photo/{id}", auth.AuthMiddleware(editPhotoHandler)).Methods("POST")

	r.Handle("/collection/move", auth.AuthMiddleware(moveCollectionHandler)).Methods("POST")
	r.Handle("/collection/uploadFile", auth.AuthMiddleware(uploadFileHandler)).Methods("POST")
	r.Handle("/collection/upload", auth.AuthMiddleware(uploadHandler)).Methods("POST")

	r.Handle("/collections", auth.AuthMiddleware(getCollectionsHandler)).Methods("GET")
	r.Handle("/collection", auth.AuthMiddleware(createCollectionHandler)).Methods("POST")
	r.Handle("/collection/{id}/photos", auth.AuthMiddleware(getCollectionPhotosHandler)).Methods("GET")
	r.Handle("/collection/{id}", auth.AuthMiddleware(getCollectionHandler)).Methods("GET")

	r.Handle("/settings/stats", auth.AuthMiddleware(statsHandler)).Methods("GET")
	r.Handle("/settings/gallery", auth.AuthMiddleware(getGallerySettings)).Methods("GET")
	r.Handle("/settings/profile", auth.AuthMiddleware(getProfileInfo)).Methods("GET")

	r.Handle("/settings/gallery", auth.AuthMiddleware(setGallerySettings)).Methods("POST")
	r.Handle("/settings/profile", auth.AuthMiddleware(setProfileInfo)).Methods("POST")

	r.Handle("/tasks/purge", auth.AuthMiddleware(purgeTaskHandler)).Methods("GET")
	r.Handle("/tasks/clear", auth.AuthMiddleware(clearTaskHandler)).Methods("GET")
	r.Handle("/tasks/backup", auth.AuthMiddleware(backupTaskHandler)).Methods("GET")
	r.Handle("/tasks/upload", auth.AuthMiddleware(uploadTaskHandler)).Methods("POST")

	return r
}
