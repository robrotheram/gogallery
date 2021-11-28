package datastore

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ahmdrz/goinsta/v2"
	"github.com/asdine/storm"
)

type DataStore struct {
	DB *storm.DB
}

var Cache *DataStore
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

func DateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
