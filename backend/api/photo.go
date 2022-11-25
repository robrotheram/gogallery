package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/backend/datastore/models"
)

// r.Handle("/api/admin/photo/{id}", auth.AuthMiddleware(getPhotoHandler)).Methods("GET")
// r.Handle("/api/admin/photo/{id}", auth.AuthMiddleware(editPhotoHandler)).Methods("POST")
// r.Handle("/api/admin/photo/{id}", auth.AuthMiddleware(deletePhotoHandler)).Methods("DELETE")

func (api *GoGalleryAPI) GetPicturesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SortByTime(api.db.Pictures.GetAll()))
}

func (api *GoGalleryAPI) GetPictureHandler(w http.ResponseWriter, r *http.Request) {
	photoID := mux.Vars(r)["id"]
	picture := api.db.Pictures.FindByID(photoID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(picture)
}

func (api *GoGalleryAPI) UpdatePictureHandler(w http.ResponseWriter, r *http.Request) {
	photoID := mux.Vars(r)["id"]
	oldPicture := api.db.Pictures.FindByID(photoID)

	var picture models.Picture
	err := json.NewDecoder(r.Body).Decode(&picture)
	if err != nil {
		NewAPIError(err).HandleError(w)
		return
	}
	if oldPicture.Name != picture.Name {
		newName := fmt.Sprintf("%s/%s%s", filepath.Dir(oldPicture.Path), picture.Name, filepath.Ext(oldPicture.Path))
		os.Rename(oldPicture.Path, newName)
		picture.Path = newName
	}

	if oldPicture.Path != picture.Path {
		os.Rename(oldPicture.Path, picture.Path)
	}

	if oldPicture.Album != picture.Album {
		api.db.Albums.MovePictureToAlbum(&picture, picture.Album)
	}
	picture.Meta.DateModified = time.Now()
	api.db.Pictures.Save(&picture)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(picture)
}

func (api *GoGalleryAPI) DeletePictureHandler(w http.ResponseWriter, r *http.Request) {
	photoID := mux.Vars(r)["id"]
	picture := api.db.Pictures.FindByID(photoID)
	api.db.Pictures.Delete(picture)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(picture)
}
