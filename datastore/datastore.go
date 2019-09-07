package datastore

import (
	"log"
	"os"
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

type Picture struct {
	Id         string `json:"id" storm:"id"`
	Name       string `json:"name"`
	Caption    string `json:"caption"`
	Path       string `json:"path"`
	FormatTime string `json:"format_time"`
	Album      string `json:"album"`
	Exif       Exif   `json:"exif"`
	PostedToIG bool   `json:"posted_to_IG"`
}

type DataStore struct {
	DB *storm.DB
}

var Cache *DataStore
var IG *Instagram

var dbVer = "1.0"

func (d *DataStore) Open(dbPath string) {
	os.MkdirAll(dbPath, os.ModePerm)
	db, err := storm.Open(dbPath + "gogallery-V" + dbVer + ".db")
	if err != nil {
		log.Fatal(err)
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
