package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/datastore"
	"github.com/robrotheram/gogallery/backend/datastore/models"
)

type Collections struct {
	CaptureDates []string                `json:"dates"`
	UploadDates  []string                `json:"uploadDates"`
	Albums       map[string]models.Album `json:"albums"`
}

type MoveCollection struct {
	Album  string           `json:"album"`
	Photos []models.Picture `json:"photos"`
}

func (api *GoGalleryAPI) moveCollectionHandler(w http.ResponseWriter, r *http.Request) {
	var moveCollection MoveCollection
	_ = json.NewDecoder(r.Body).Decode(&moveCollection)
	for _, photo := range moveCollection.Photos {
		oldPicture := api.db.Pictures.FindByID(photo.Id)
		if oldPicture.Album != moveCollection.Album {
			api.db.Albums.MovePictureToAlbum(&photo, moveCollection.Album)
			photo.Meta.DateModified = time.Now()
			api.db.Pictures.Save(&photo)
		}
	}

}

func (api *GoGalleryAPI) updateCollectionHandler(w http.ResponseWriter, r *http.Request) {
	albumID := mux.Vars(r)["id"]
	oldAlbum := api.db.Albums.FindByID(albumID)

	var album models.Album
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

	api.db.Albums.Save(&oldAlbum)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(oldAlbum)
}

func (api *GoGalleryAPI) createCollectionHandler(w http.ResponseWriter, r *http.Request) {
	var album models.Album
	_ = json.NewDecoder(r.Body).Decode(&album)

	path := ""
	if album.Id != "" {
		newAlbum := api.db.Albums.FindByID(album.Id)
		path = fmt.Sprintf("%s/%s/%s", newAlbum.ParenetPath, newAlbum.Name, album.Name)
	} else {
		path = fmt.Sprintf("%s/%s", api.config.Gallery.Basepath, album.Name)
	}

	album.Id = config.GetMD5Hash(path)
	album.ParenetPath = filepath.Dir(path)
	album.ModTime = time.Now()
	album.Children = make(map[string]models.Album)

	fmt.Println(album)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}

	api.db.Albums.Save(&album)
}

func (api *GoGalleryAPI) getCollectionPhotosHandler(w http.ResponseWriter, r *http.Request) {
	albumID := mux.Vars(r)["id"]
	pictures := api.db.Pictures.FindManyFeild("Album", albumID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pictures)
}

func (api *GoGalleryAPI) getCollectionHandler(w http.ResponseWriter, r *http.Request) {
	albumID := mux.Vars(r)["id"]
	album := api.db.Albums.FindByID(albumID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(album)
}

func (api *GoGalleryAPI) getCollectionsHandler(w http.ResponseWriter, r *http.Request) {

	pics := api.db.Pictures.GetAll()
	albms := api.db.Albums.GetAll()
	newalbms := datastore.SliceToTree(albms, api.config.Gallery.Basepath)
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
}
