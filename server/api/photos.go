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
	templateengine "github.com/robrotheram/gogallery/templateEngine"
)

var editPhotoHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	photoID := mux.Vars(r)["id"]
	var picture datastore.Picture
	err := json.NewDecoder(r.Body).Decode(&picture)

	if err != nil {
		NewAPIError(err).HandleError(w)
		return
	}

	oldPicture, err := datastore.GetPictureByID(photoID)

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
		picture.MoveToAlbum(picture.Album)
	}

	picture.Meta.DateModified = time.Now()
	picture.Save()
	templateengine.InvalidCache()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(picture)
})

var deletePhotoHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	photoID := mux.Vars(r)["id"]
	picture, err := datastore.GetPictureByID(photoID)
	if err != nil {
		NewAPIError(err).HandleError(w)
		return
	}
	err = picture.Delete()
	err = datastore.Cache.DB.DeleteStruct(&picture)
	if err != nil {
		NewAPIError(err).HandleError(w)
		return
	}

	templateengine.InvalidCache()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(picture)
})

var getPhotoHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	photoID := mux.Vars(r)["id"]
	picture, err := datastore.GetPictureByID(photoID)
	if err != nil {
		NewAPIError(err).HandleError(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(picture)
})

var getAllAdminPhotosHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	pics := datastore.GetFilteredPictures(true)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pics)
})

var getAllPhotosHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	filterPics := datastore.GetFilteredPictures(false)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filterPics)
})

var getLatestCollectionsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	pics := datastore.GetFilteredPictures(false)
	latests := pics[0].Exif.DateTaken
	for _, p := range pics {
		if p.Exif.DateTaken.After(latests) {
			latests = p.Exif.DateTaken
		}
	}
	w.Write([]byte(latests.Format("2006-01-02")))
})

var getByDatePhotosHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	date := mux.Vars(r)["date"]
	yourDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		yourDate, err = time.Parse("01-02-2006", date)
	}
	if err != nil {
		http.Error(w, "Invalid date", http.StatusBadRequest)
	}
	latests := datastore.GetPhotosByDate(yourDate)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(latests)
})

var CaptionHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	photoID := mux.Vars(r)["id"]
	photo, err := datastore.GetPictureByID(photoID)
	if err != nil {
		http.Error(w, "Photo Not Found", http.StatusBadRequest)
		return
	}
	caption, err := datastore.GetCaptions(&photo)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get Caption: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(caption)
})

func DateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
