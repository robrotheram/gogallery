package datastore

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ahmdrz/goinsta/v2"
	"github.com/asdine/storm"
	Config "github.com/robrotheram/gogallery/config"
)

type Album struct {
	Id          string           `json:"id" storm:"id"`
	Name        string           `json:"name"`
	ModTime     time.Time        `json:"mod_time"`
	Parent      string           `json:"parent"`
	ParenetPath string           `json:"parentPath,omitempty"`
	ProfileID   string           `json:"profile_image"`
	Images      []Picture        `json:"images"`
	Children    map[string]Album `json:"children"`
	GPS         GPS              `json: gps`
}

type AlbumStrcure = map[string]Album
type UploadCollection struct {
	Album  string   `json:"album"`
	Photos []string `json:"photos"`
}

type Collections struct {
	CaptureDates []string         `json:"dates"`
	UploadDates  []string         `json:"uploadDates"`
	Albums       map[string]Album `json:"albums"`
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
	GPS          GPS       `json: gps`
}

type GPS struct {
	Lat float64 `json:"latitude"`
	Lng float64 `json:"longitude"`
}

type PictureMeta struct {
	PostedToIG   bool      `json:"posted_to_IG,omitempty"`
	Visibility   string    `json:"visibility,omitempty"`
	DateAdded    time.Time `json: date_added,omitempty`
	DateModified time.Time `json: date_modified,omitempty`
}

type Picture struct {
	Id         string      `json:"id" storm:"id"`
	Name       string      `json:"name"`
	Caption    string      `json:"caption"`
	Path       string      `json:"path,omitempty"`
	FormatTime string      `json:"format_time"`
	Album      string      `json:"album"`
	AlbumName  string      `json:"album_name"`
	Exif       Exif        `json:"exif"`
	Meta       PictureMeta `json:"meta,omitempty"`
	RootPath   string      `json:"root_path,omitempty"`
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
	var album Album
	Cache.DB.One("Id", newAlbum, &album)

	newName := fmt.Sprintf("%s/%s/%s%s", album.ParenetPath, album.Name, picture.Name, filepath.Ext(oldPath))
	picture.Path = newName
	picture.Album = newAlbum

	fmt.Println(newName)
	err := os.Rename(oldPath, newName)
	if err != nil {
		fmt.Println(err)
	}
}

func (picture *Picture) Delete() {
	err := os.Remove(picture.Path)
	if err != nil {
		fmt.Println(err)
	}
}

func SliceToTree(albms []Album, basepath string) map[string]Album {
	newalbms := make(map[string]Album)
	sort.Slice(albms, func(i, j int) bool {
		return albms[i].ParenetPath < albms[j].ParenetPath
	})
	for _, ab := range albms {
		if ab.ParenetPath == basepath {
			ab.ParenetPath = ""
			newalbms[ab.Name] = ab
		}
	}
	for _, ab := range albms {
		if (ab.ParenetPath != basepath) && (ab.Id != Config.GetMD5Hash(basepath)) {
			s := strings.Split(strings.Replace(ab.ParenetPath, basepath, "", 1), "/")
			copy(s, s[1:])
			s = s[:len(s)-1]
			pth := basepath
			var alb Album
			for i, p := range s {
				if i == 0 {
					alb = newalbms[p]
				} else {
					alb = alb.Children[p]
				}
				pth = path.Join(pth, p)
				if i == len(s)-1 {
					if alb.Children != nil {
						ab.ParenetPath = ""
						alb.Children[ab.Name] = ab
					}
				}
			}
		}
	}
	return newalbms
}

func (a *Album) Update(alb Album) {

	if a.Name != alb.Name && alb.Name != "" {
		a.Name = alb.Name
	}
	if a.Parent != alb.Parent && alb.Parent != "" {
		a.Parent = alb.Parent
	}
	if a.ParenetPath != alb.ParenetPath && alb.ParenetPath != "" {
		a.ParenetPath = alb.ParenetPath
	}
	if (a.ProfileID != alb.ProfileID) && (alb.ProfileID != "") {
		a.ProfileID = alb.ProfileID
	}
	if a.Children == nil {
		a.Children = make(map[string]Album)
	}
	if a.Id == "" {
		a.Id = alb.Id
	}

}
