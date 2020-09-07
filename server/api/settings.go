package api

import (
	"encoding/json"
	"github.com/robrotheram/gogallery/config"
	"net/http"
)

var statsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MakeStats())
})

var getProfileInfo = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Config.About)
})

var setProfileInfo = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var about = config.AboutConfiguration{}
	_ = json.NewDecoder(r.Body).Decode(&about)
	about.Save()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Config.About)
})

var getGallerySettings = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Config.Gallery)
})

var getPublicGallerySettings = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	config := map[string]string{
		"name": Config.Gallery.Name,
	}
	json.NewEncoder(w).Encode(config)
})

var setGallerySettings = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var gallery = config.GalleryConfiguration{}
	_ = json.NewDecoder(r.Body).Decode(&gallery)
	gallery.Save()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Config.Gallery)
})
