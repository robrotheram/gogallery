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

	r.Handle("/api/albums", (getAllCollectionsHandler)).Methods("GET")
	r.Handle("/api/photos", (getAllPhotosHandler)).Methods("GET")
	r.Handle("/api/profile", (getProfileInfo)).Methods("GET")

	r.Handle("/api/admin/photos", auth.AuthMiddleware(getAllAdminPhotosHandler))
	r.Handle("/api/admin/photo/{id}", auth.AuthMiddleware(getPhotoHandler)).Methods("GET")
	r.Handle("/api/admin/photo/{id}", auth.AuthMiddleware(editPhotoHandler)).Methods("POST")

	r.Handle("/api/admin/collection/move", auth.AuthMiddleware(moveCollectionHandler)).Methods("POST")
	r.Handle("/api/admin/collection/uploadFile", auth.AuthMiddleware(uploadFileHandler)).Methods("POST")
	r.Handle("/api/admin/collection/upload", auth.AuthMiddleware(uploadHandler)).Methods("POST")

	r.Handle("/api/admin/collections", auth.AuthMiddleware(getCollectionsHandler)).Methods("GET")
	r.Handle("/api/admin/collection", auth.AuthMiddleware(createCollectionHandler)).Methods("POST")
	r.Handle("/api/admin/collection/{id}/photos", auth.AuthMiddleware(getCollectionPhotosHandler)).Methods("GET")
	r.Handle("/api/admin/collection/{id}", auth.AuthMiddleware(getCollectionHandler)).Methods("GET")

	r.Handle("/api/admin/settings/stats", auth.AuthMiddleware(statsHandler)).Methods("GET")
	r.Handle("/api/admin/settings/gallery", auth.AuthMiddleware(getGallerySettings)).Methods("GET")
	r.Handle("/api/admin/settings/profile", auth.AuthMiddleware(getProfileInfo)).Methods("GET")

	r.Handle("/api/admin/settings/gallery", auth.AuthMiddleware(setGallerySettings)).Methods("POST")
	r.Handle("/api/admin/settings/profile", auth.AuthMiddleware(setProfileInfo)).Methods("POST")

	r.Handle("/api/admin/tasks/purge", auth.AuthMiddleware(purgeTaskHandler)).Methods("GET")
	r.Handle("/api/admin/tasks/clear", auth.AuthMiddleware(clearTaskHandler)).Methods("GET")
	r.Handle("/api/admin/tasks/backup", auth.AuthMiddleware(backupTaskHandler)).Methods("GET")
	r.Handle("/api/admin/tasks/upload", auth.AuthMiddleware(uploadTaskHandler)).Methods("POST")

	return r
}
