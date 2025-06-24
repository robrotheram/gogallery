package api

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/robrotheram/gogallery/backend/config"
)

type Stats struct {
	Photos int
	Albums int
	Rubish int
}

func (api *GoGalleryAPI) statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stats := Stats{0, 0, 0}

	photos, _ := api.Pictures.GetAll()
	albums, _ := api.Albums.GetAll()

	stats.Photos = len(photos)
	stats.Albums = len(albums)
	files, _ := os.ReadDir(config.Config.Gallery.Basepath + "/rubish")
	stats.Rubish = len(files)
	json.NewEncoder(w).Encode(stats)
}

func (api *GoGalleryAPI) getProfileInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.config.About)
}

func (api *GoGalleryAPI) setProfileInfo(w http.ResponseWriter, r *http.Request) {
	var about = config.AboutConfiguration{}
	_ = json.NewDecoder(r.Body).Decode(&about)
	about.Save()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.config.About)
}

func (api *GoGalleryAPI) getGallerySettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.config.Gallery)
}

func (api *GoGalleryAPI) getDeploymentSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.config.Deploy)
}

func (api *GoGalleryAPI) getPublicGallerySettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	config := map[string]string{
		"name": api.config.Gallery.Name,
	}
	json.NewEncoder(w).Encode(config)
}

func (api *GoGalleryAPI) setGallerySettings(w http.ResponseWriter, r *http.Request) {
	var gallery = config.GalleryConfiguration{}
	_ = json.NewDecoder(r.Body).Decode(&gallery)
	gallery.Save()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.config.Gallery)
}

func (api *GoGalleryAPI) setDeploymentSettings(w http.ResponseWriter, r *http.Request) {
	var deploy = config.DeployConfig{}
	_ = json.NewDecoder(r.Body).Decode(&deploy)
	deploy.Save()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.config.Deploy)
}
