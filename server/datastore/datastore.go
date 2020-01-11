package datastore

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ahmdrz/goinsta/v2"
	"github.com/asdine/storm"
)

type Album struct {
	Id         string    `json:"id" storm:"id"`
	Name       string    `json:"name"`
	ModTime    time.Time `json:"mod_time"`
	Parent     string    `json:"parent"`
	ProfileIMG *Picture  `json:"profile_image"`
	Images     []Picture `json:"images"`
	Key        string    `json:"key"`
}

type UploadCollection struct {
	Album  string   `json:"album"`
	Photos []string `json:"photos"`
}

type Collections struct {
	CaptureDates []string `json:"dates"`
	UploadDates  []string `json:"uploadDates"`
	Albums       []Album  `json:"albums"`
}

type MoveCollection struct {
	Album  string    `json:"album"`
	Photos []Picture `json:"photos"`
}

type Directory struct {
	Album    Album        `json:"album"`
	Children []*Directory `json:"children"`
}

type Exif struct {
	FStop        float64   `json:"f_stop"`
	FocalLength  float64   `json:"focal_length"`
	ShutterSpeed string    `json:"shutter_speed"`
	ISO          string    `json:"iso"`
	Dimension    string    `json:"dimension"`
	Camera       string    `json:"camera"`
	LensModel    string    `json: lens_model`
	DateTaken    time.Time `json: date_taken`
}

type PictureMeta struct {
	PostedToIG   bool      `json:"posted_to_IG"`
	Visibility   string    `json:"visibility"`
	DateAdded    time.Time `json: date_added`
	DateModified time.Time `json: date_modified`
}

type Picture struct {
	Id         string      `json:"id" storm:"id"`
	Name       string      `json:"name"`
	Caption    string      `json:"caption"`
	Path       string      `json:"path"`
	FormatTime string      `json:"format_time"`
	Album      string      `json:"album"`
	Exif       Exif        `json:"exif"`
	Meta       PictureMeta `json:"meta"`
	RootPath   string
}

type User struct {
	ID       string `json:"id,omitempty" storm:"id"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
	Token    string `json:"token,omitempty"`
}

type DataStore struct {
	DB *storm.DB
}

var Cache *DataStore

//var IG *Instagram

var dbVer = "1.0"

func (d *DataStore) Open(dbPath string) {
	os.MkdirAll(dbPath, os.ModePerm)
	path := fmt.Sprintf("%sgogallery-V%s.db", dbPath, dbVer)
	db, err := storm.Open(path)
	if err != nil {
		log.Fatalf("Unable to open db at: %s \n Error: %v", path, err)
	}
	d.DB = db
}

func (d *DataStore) Close() {
	d.DB.Close()
}

func (d *DataStore) RestDB() {
	d.DB.Drop(Picture{})
	d.DB.Drop(Album{})
	d.DB.Drop(goinsta.Item{})
}

func (picture *Picture) MoveToAlbum(newAlbum string) {
	oldPath := picture.Path
	basepath := filepath.Dir(filepath.Dir(oldPath))
	if filepath.Dir(oldPath) == picture.RootPath {
		basepath = (filepath.Dir(oldPath))
	}
	newName := fmt.Sprintf("%s/%s/%s%s", basepath, newAlbum, picture.Name, filepath.Ext(oldPath))
	picture.Path = newName
	picture.Album = newAlbum

	fmt.Println(newName)
	err := os.Rename(oldPath, newName)
	if err != nil {
		fmt.Println(err)
	}
}
