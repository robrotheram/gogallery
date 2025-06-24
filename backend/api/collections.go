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
)

type Collections struct {
	CaptureDates []string               `json:"dates"`
	UploadDates  []string               `json:"uploadDates"`
	Albums       datastore.AlbumStrcure `json:"albums"`
}

type MoveCollection struct {
	Album  string              `json:"album"`
	Photos []datastore.Picture `json:"photos"`
}

func (api *GoGalleryAPI) moveCollectionHandler(w http.ResponseWriter, r *http.Request) {
	var moveCollection MoveCollection
	_ = json.NewDecoder(r.Body).Decode(&moveCollection)
	for _, photo := range moveCollection.Photos {
		oldPicture, _ := api.Pictures.FindById(photo.Id)
		if oldPicture.Album != moveCollection.Album {
			api.Albums.MovePictureToAlbum(photo, moveCollection.Album)
			photo.DateModified = time.Now()
			api.Pictures.Save(photo)
		}
	}

}

func (api *GoGalleryAPI) updateCollectionHandler(w http.ResponseWriter, r *http.Request) {
	albumId := mux.Vars(r)["Id"]
	oldAlbum, _ := api.Albums.FindById(albumId)

	var album datastore.Album
	_ = json.NewDecoder(r.Body).Decode(&album)

	if oldAlbum.Name != album.Name {
		oldPath := fmt.Sprintf("%s/%s", filepath.Dir(oldAlbum.ParentPath), oldAlbum.Name)
		newPath := fmt.Sprintf("%s/%s", filepath.Dir(oldAlbum.ParentPath), album.Name)
		os.Rename(oldPath, newPath)
		oldAlbum.Name = album.Name
	}

	if oldAlbum.ProfileId != album.ProfileId {
		oldAlbum.ProfileId = album.ProfileId
	}

	api.Albums.Update(oldAlbum.Id, oldAlbum)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(oldAlbum)
}

func (api *GoGalleryAPI) createCollectionHandler(w http.ResponseWriter, r *http.Request) {
	var album datastore.Album
	_ = json.NewDecoder(r.Body).Decode(&album)

	path := ""
	if album.Id != "" {
		newAlbum, _ := api.Albums.FindById(album.Id)
		path = fmt.Sprintf("%s/%s/%s", newAlbum.ParentPath, newAlbum.Name, album.Name)
	} else {
		path = fmt.Sprintf("%s/%s", api.config.Gallery.Basepath, album.Name)
	}

	album.Id = config.GetMD5Hash(path)
	album.Parent = filepath.Dir(path)
	album.ModTime = time.Now()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}
	api.Albums.Save(album)
}

func (api *GoGalleryAPI) getCollectionPhotosHandler(w http.ResponseWriter, r *http.Request) {
	albumId := mux.Vars(r)["Id"]
	pictures, _ := api.Pictures.FindByField("Album", albumId)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pictures)
}

func (api *GoGalleryAPI) getCollectionHandler(w http.ResponseWriter, r *http.Request) {
	albumId := mux.Vars(r)["Id"]
	album, _ := api.Albums.FindById(albumId)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(album)
}

func (api *GoGalleryAPI) getCollectionsHandler(w http.ResponseWriter, r *http.Request) {

	pics, _ := api.Pictures.GetAll()
	albms, _ := api.Albums.GetAll()
	newalbms := datastore.SliceToTree(albms, api.config.Gallery.Basepath)
	dates := []string{}
	uploadDates := []string{}
	for _, pic := range pics {
		_date := pic.DateTaken.Format("2006-01-02")
		_uploadDate := pic.DateAdded.Format("2006-01-02 15:04")

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
