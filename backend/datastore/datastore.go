package datastore

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/asdine/storm"
	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/datastore/models"
)

type CRUD interface {
	Save(any)
	Delete(any)
	Update(any)
	GetById(string) []any
	GetByField(string string) []any
	GetAll() []any
}

type DataStore struct {
	db       *storm.DB
	Pictures *PictureCollection
	Albums   *AlumnCollectioins
}

func Open(dbPath string) *DataStore {
	os.MkdirAll(dbPath, os.ModePerm)
	path := filepath.Join(dbPath, "gogallery.db")
	db, err := storm.Open(path)
	if err != nil {
		log.Fatalf("Unable to open db at: %s \n Error: %v", path, err)
	}
	return &DataStore{
		db:       db,
		Pictures: &PictureCollection{DB: db},
		Albums:   &AlumnCollectioins{DB: db},
	}
}

func (d *DataStore) Close() {
	d.db.Close()
}

func (d *DataStore) RestDB() {
	d.db.Drop(models.Picture{})
	d.db.Drop(models.Album{})
}

func (d *DataStore) ScanPath(path string) error {
	rubishPath := fmt.Sprintf("%s/%s", gConfig.Basepath, "rubish")
	if _, err := os.Stat(rubishPath); os.IsNotExist(err) {
		os.Mkdir(rubishPath, 0755)
	}
	if !contains(gConfig.AlbumBlacklist, "rubish") {
		gConfig.AlbumBlacklist = append(gConfig.AlbumBlacklist, "rubish")
	}

	log.Println("Scanning Folders at:" + path)
	IsScanning = true

	absRoot, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if checkEXT(path) && !info.IsDir() {
			albumName := filepath.Base(filepath.Dir(path))
			picName := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
			if !IsAlbumInBlacklist(albumName) && !IsPictureInBlacklist(picName) {
				p := models.Picture{
					Id:        config.GetMD5Hash(path),
					Name:      picName,
					Path:      path,
					Ext:       filepath.Ext(path),
					Album:     config.GetMD5Hash(filepath.Dir(path)),
					AlbumName: albumName,
					Exif:      models.Exif{},
					RootPath:  gConfig.Basepath,
					Meta: models.PictureMeta{
						PostedToIG:   false,
						Visibility:   "PUBLIC",
						DateAdded:    time.Now(),
						DateModified: time.Now()}}
				p.CreateExif()
				if !d.Pictures.Exist(p.Id) {
					d.Pictures.Save(&p)
				}
				d.Albums.UpdateField(config.GetMD5Hash(filepath.Dir(path)), "ProfileID", p.Id)
			}
		}

		if info.IsDir() {
			if !IsAlbumInBlacklist(info.Name()) {
				if filepath.Base(filepath.Dir(path)) != gConfig.Basepath {
					info := fileInfoFromInterface(info)
					d.Albums.Update(&models.Album{
						Id:          config.GetMD5Hash(path),
						Name:        info.Name,
						ModTime:     info.ModTime,
						Parent:      filepath.Base(filepath.Dir(path)),
						ParenetPath: (filepath.Dir(path))})
				}
			}
		}
		return nil
	}
	err = filepath.Walk(absRoot, walkFunc)
	log.Println("Scanning Complete")
	IsScanning = false
	return err
}
