package serve

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/datastore"
	templateengine "github.com/robrotheram/gogallery/backend/templateEngine"
	"github.com/robrotheram/gogallery/backend/worker"
)

type backup struct {
	Albums   []datastore.Album    `json:"albums"`
	Pictures []datastore.Picture  `json:"pictures"`
	Config   config.Configuration `json:"config"`
}

var purgeTaskHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	log.Println("DeletingDB")
	datastore.Cache.RestDB()
})

var rescanTaskHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	go func() {
		worker.ScanPath(Config.Gallery.Basepath)
	}()
})

var clearTaskHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	worker.RemoveContents("cache")
})

var uploadTaskHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		datastore.Cache.DB.Save(&p)
	}
	for _, a := range bk.Albums {
		datastore.Cache.DB.Save(&a)
	}
	bk.Config.About.Save()
	bk.Config.Gallery.Save()
})

var backupTaskHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	bk := backup{}
	datastore.Cache.DB.All(&bk.Pictures)
	datastore.Cache.DB.All(&bk.Albums)
	bk.Config = *Config

	w.Header().Set("Content-Disposition", "attachment; filename=Gallery-Backup.json")
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	json.NewEncoder(w).Encode(bk)
})

var templateInvalidateHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	templateengine.InvalidCache()
})