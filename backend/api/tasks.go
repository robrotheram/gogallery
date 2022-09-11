package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/datastore/models"
	"github.com/robrotheram/gogallery/backend/deploy"
	"github.com/robrotheram/gogallery/backend/pipeline"
)

type backup struct {
	Albums   []models.Album       `json:"albums"`
	Pictures []models.Picture     `json:"pictures"`
	Config   config.Configuration `json:"config"`
}

func (api *GoGalleryAPI) purgeTaskHandler(w http.ResponseWriter, r *http.Request) {
	pipeline.NewRenderPipeline(&api.config.Gallery, api.db, api.monitor).DeleteSite()
	fmt.Fprintf(w, "Deleted Site")
}

func (api *GoGalleryAPI) getTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.monitor.GetTasks())
}

func (api *GoGalleryAPI) rescanTaskHandler(w http.ResponseWriter, r *http.Request) {
	go func() {
		api.db.ScanPath(api.config.Gallery.Basepath)
	}()
}

func (api *GoGalleryAPI) buildTaskHandler(w http.ResponseWriter, r *http.Request) {
	go pipeline.NewRenderPipeline(&api.config.Gallery, api.db, api.monitor).BuildSite()
	fmt.Fprintf(w, "Build task started")
}

func (api *GoGalleryAPI) deployTaskHandler(w http.ResponseWriter, r *http.Request) {
	go deploy.DeploySite(*api.config, api.monitor.NewTask("netify deploy"))
	fmt.Fprintf(w, "Deploy task started")
}

func (api *GoGalleryAPI) uploadTaskHandler(w http.ResponseWriter, r *http.Request) {
	bk := backup{}
	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile("file")
	defer file.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		fmt.Println(err)
		return
	}
	json.Unmarshal(buf.Bytes(), &bk)
	for _, p := range bk.Pictures {
		api.db.Pictures.Save(&p)
	}
	for _, a := range bk.Albums {
		api.db.Albums.Save(&a)
	}
	bk.Config.About.Save()
	bk.Config.Gallery.Save()
}

func (api *GoGalleryAPI) backupTaskHandler(w http.ResponseWriter, r *http.Request) {
	bk := backup{}
	bk.Albums = api.db.Albums.GetAll()
	bk.Pictures = api.db.Pictures.GetAll()
	bk.Config = *api.config

	w.Header().Set("Content-Disposition", "attachment; filename=Gallery-Backup.json")
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	json.NewEncoder(w).Encode(bk)
}
