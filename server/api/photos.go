package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/datastore"
)

var editPhotoHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	photoID := mux.Vars(r)["id"]
	var oldPicture datastore.Picture
	var picture datastore.Picture
	datastore.Cache.DB.One("Id", photoID, &oldPicture)
	_ = json.NewDecoder(r.Body).Decode(&picture)

	if oldPicture.Name != picture.Name {
		newName := fmt.Sprintf("%s/%s%s", filepath.Dir(oldPicture.Path), picture.Name, filepath.Ext(oldPicture.Path))
		os.Rename(oldPicture.Path, newName)
		picture.Path = newName
	}
	if oldPicture.Album != picture.Album {
		picture.MoveToAlbum(picture.Album)
	}

	picture.Meta.DateModified = time.Now()
	datastore.Cache.DB.Save(&picture)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(picture)
})

var getPhotoHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	photoID := mux.Vars(r)["id"]
	var picture datastore.Picture
	datastore.Cache.DB.One("Id", photoID, &picture)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(picture)
})

var getAllPhotosHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var pics []datastore.Picture
	var filterPics []datastore.Picture
	datastore.Cache.DB.All(&pics)
	for _, pic := range pics {
		if !datastore.IsAlbumInBlacklist(pic.Album) {
			filterPics = append(filterPics, pic)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filterPics)
})
