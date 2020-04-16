package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
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

var deletePhotoHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	photoID := mux.Vars(r)["id"]
	var oldPicture datastore.Picture
	datastore.Cache.DB.One("Id", photoID, &oldPicture)
	datastore.Cache.DB.DeleteStruct(&oldPicture)
	oldPicture.Delete()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(oldPicture)
})

var getPhotoHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	photoID := mux.Vars(r)["id"]
	var picture datastore.Picture
	datastore.Cache.DB.One("Id", photoID, &picture)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(picture)
})

var getAllAdminPhotosHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var pics []datastore.Picture
	var filterPics []datastore.Picture
	datastore.Cache.DB.All(&pics)
	for _, pic := range pics {
		if !datastore.IsAlbumInBlacklist(pic.Album) {
			filterPics = append(filterPics, pic)
		}
	}
	sort.Slice(filterPics, func(i, j int) bool {
		return filterPics[i].Exif.DateTaken.Sub(filterPics[j].Exif.DateTaken) > 0
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filterPics)
})

var getAllPhotosHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var pics []datastore.Picture
	var filterPics []datastore.Picture
	datastore.Cache.DB.All(&pics)
	for _, pic := range pics {
		if !datastore.IsAlbumInBlacklist(pic.Album) {
			if pic.Meta.Visibility == "PUBLIC" {
				cleanpic := datastore.Picture{
					Id:         pic.Id,
					Name:       pic.Name,
					Caption:    pic.Caption,
					Album:      pic.Album,
					FormatTime: pic.Exif.DateTaken.Format("01-02-2006 15:04:05"),
					Exif:       pic.Exif,
					Meta:       pic.Meta,
				}
				filterPics = append(filterPics, cleanpic)
			}

		}
	}
	sort.Slice(filterPics, func(i, j int) bool {
		return filterPics[i].Exif.DateTaken.Sub(filterPics[j].Exif.DateTaken) > 0
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filterPics)
})
