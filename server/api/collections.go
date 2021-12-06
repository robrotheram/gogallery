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
	templateengine "github.com/robrotheram/gogallery/templateEngine"
)

type Collections struct {
	CaptureDates []string                   `json:"dates"`
	UploadDates  []string                   `json:"uploadDates"`
	Albums       map[string]datastore.Album `json:"albums"`
}

type MoveCollection struct {
	Album  string              `json:"album"`
	Photos []datastore.Picture `json:"photos"`
}

var moveCollectionHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var moveCollection MoveCollection
	_ = json.NewDecoder(r.Body).Decode(&moveCollection)
	for _, photo := range moveCollection.Photos {
		oldPicture, err := datastore.GetPictureByID(photo.Id)
		if err != nil {
			break
		}
		if oldPicture.Album != moveCollection.Album {
			photo.MoveToAlbum(moveCollection.Album)
			photo.Meta.DateModified = time.Now()
			photo.Save()
		}
	}
	templateengine.InvalidCache()
})

var updateCollectionHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	albumID := mux.Vars(r)["id"]
	oldAlbum, _ := datastore.GetAlbumByID(albumID)

	var album datastore.Album
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

	if oldAlbum.GPS.Lat != album.GPS.Lat || oldAlbum.GPS.Lng != album.GPS.Lng {
		oldAlbum.GPS = album.GPS
	}

	oldAlbum.Save()
	templateengine.InvalidCache()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(oldAlbum)
})

var createCollectionHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var album datastore.Album
	_ = json.NewDecoder(r.Body).Decode(&album)
	newAlbum, _ := datastore.GetAlbumByID(album.Id)

	path := fmt.Sprintf("%s/%s/%s", newAlbum.ParenetPath, newAlbum.Name, album.Name)

	album.Id = config.GetMD5Hash(path)
	album.ParenetPath = filepath.Dir(path)
	album.ModTime = time.Now()
	album.Children = make(map[string]datastore.Album)

	fmt.Println(album)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}
	templateengine.InvalidCache()
	album.Save()
})

var getAllCollectionsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	newalbms := datastore.GetAlbumStructure(Config.Gallery)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newalbms)
})

var getCollectionHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	albumID := mux.Vars(r)["id"]
	album, _ := datastore.GetAlbumByID(albumID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(album)
})

var getCollectionPhotosHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	albumID := mux.Vars(r)["id"]
	pictures := datastore.GetPicturesByAlbumID(albumID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pictures)
})

var getCollectionsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	pics := datastore.GetPictures()
	albms := datastore.GetAlbums()
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
	json.NewEncoder(w).Encode(Collections{CaptureDates: dates, UploadDates: uploadDates, Albums: newalbms})
})
