package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/backend/datastore"
)

func (api *GoGalleryAPI) GetPicturesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	pics, _ := api.Pictures.GetAll()
	json.NewEncoder(w).Encode(pics)
}

func (api *GoGalleryAPI) GetPictureHandler(w http.ResponseWriter, r *http.Request) {
	photoID := mux.Vars(r)["id"]
	picture, _ := api.Pictures.FindById(photoID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(picture)
}

func (api *GoGalleryAPI) UpdatePictureHandler(w http.ResponseWriter, r *http.Request) {
	photoID := mux.Vars(r)["id"]
	oldPicture, _ := api.Pictures.FindById(photoID)

	var picture datastore.Picture
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
		api.Albums.MovePictureToAlbum(picture, picture.Album)
	}
	picture.DateModified = time.Now()
	api.Pictures.Update(photoID, picture)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(picture)
}

func (api *GoGalleryAPI) DeletePictureHandler(w http.ResponseWriter, r *http.Request) {
	photoID := mux.Vars(r)["id"]
	picture, _ := api.Pictures.FindById(photoID)
	api.Pictures.Delete(picture)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(picture)
}
