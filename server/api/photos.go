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
				var album datastore.Album
				datastore.Cache.DB.One("Id", pic.Album, &album)
				cleanpic := datastore.Picture{
					Id:         pic.Id,
					Name:       pic.Name,
					Caption:    pic.Caption,
					Album:      pic.Album,
					AlbumName:  album.Name,
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

var getLatestCollectionsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var pics []datastore.Picture
	datastore.Cache.DB.All(&pics)
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
	var pics []datastore.Picture
	datastore.Cache.DB.All(&pics)
	latests := []datastore.Picture{}
	for _, p := range pics {
		if DateEqual(p.Exif.DateTaken, yourDate) {
			latests = append(latests, p)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(latests)
})

func DateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
