package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
)

var moveCollectionHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var moveCollection datastore.MoveCollection
	_ = json.NewDecoder(r.Body).Decode(&moveCollection)
	for _, photo := range moveCollection.Photos {
		var oldPicture datastore.Picture
		datastore.Cache.DB.One("Id", photo.Id, &oldPicture)
		if oldPicture.Album != moveCollection.Album {
			photo.MoveToAlbum(moveCollection.Album)
			photo.Meta.DateModified = time.Now()
			datastore.Cache.DB.Save(&photo)
		}
	}
})

var updateCollectionHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	albumID := mux.Vars(r)["id"]
	var oldAlbum datastore.Album
	var album datastore.Album
	datastore.Cache.DB.One("Id", albumID, &oldAlbum)
	_ = json.NewDecoder(r.Body).Decode(&album)

	if oldAlbum.Name != album.Name {
		oldPath := fmt.Sprintf("%s/%s", filepath.Dir(oldAlbum.ParenetPath), oldAlbum.Name)
		newPath := fmt.Sprintf("%s/%s", filepath.Dir(oldAlbum.ParenetPath), album.Name)
		os.Rename(oldPath, newPath)
		oldAlbum.Name = album.Name
	}

	if oldAlbum.ProfileID != album.ProfileID {
		oldAlbum.ProfileID = album.ProfileID
	}

	datastore.Cache.DB.Save(&oldAlbum)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(oldAlbum)
})

var createCollectionHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var album datastore.Album
	var newAlbum datastore.Album

	_ = json.NewDecoder(r.Body).Decode(&album)
	datastore.Cache.DB.One("Id", album.Id, &newAlbum)

	path := fmt.Sprintf("%s/%s/%s", newAlbum.ParenetPath, newAlbum.Name, album.Name)

	album.Id = config.GetMD5Hash(path)
	album.ParenetPath = filepath.Dir(path)
	album.ModTime = time.Now()
	album.Children = make(map[string]datastore.Album)

	fmt.Println(album)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}

	datastore.Cache.DB.Save(&album)
})

var getAllCollectionsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var albms []datastore.Album
	datastore.Cache.DB.All(&albms)
	newalbms := datastore.SliceToTree(albms, Config.Gallery.Basepath)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newalbms)
})

var getCollectionHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	albumID := mux.Vars(r)["id"]
	var album datastore.Album
	datastore.Cache.DB.One("Name", albumID, &album)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(album)
})

var getCollectionPhotosHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	albumID := mux.Vars(r)["id"]
	var pictures []datastore.Picture
	datastore.Cache.DB.Find("Album", albumID, &pictures)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pictures)
})

var getCollectionsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var pics []datastore.Picture
	datastore.Cache.DB.All(&pics)
	var albms []datastore.Album
	datastore.Cache.DB.All(&albms)
	newalbms := datastore.SliceToTree(albms, Config.Gallery.Basepath)
	dates := []string{}
	uploadDates := []string{}
	for _, pic := range pics {
		_date := pic.Exif.DateTaken.Format("2006-01-02")
		_uploadDate := pic.Meta.DateAdded.Format("2006-01-02 15:04")

		found := false
		uploadFound := false

		for _, date := range dates {
			if date == _date {
				found = true
			}
		}
		for _, date := range uploadDates {
			if date == _uploadDate {
				uploadFound = true
			}
		}
		if !found {
			dates = append(dates, _date)
		}
		if !uploadFound {
			uploadDates = append(uploadDates, _uploadDate)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(datastore.Collections{CaptureDates: dates, UploadDates: uploadDates, Albums: newalbms})
})
