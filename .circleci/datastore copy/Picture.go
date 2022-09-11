package datastore

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/evanoberholster/imagemeta"
	"github.com/robrotheram/gogallery/backend/config"
)

type Picture struct {
	Id         string      `json:"id" storm:"id"`
	Name       string      `json:"name"`
	Caption    string      `json:"caption"`
	Path       string      `json:"path,omitempty"`
	Ext        string      `json:"extention,omitempty"`
	FormatTime string      `json:"format_time"`
	Album      string      `json:"album"`
	AlbumName  string      `json:"album_name"`
	Exif       Exif        `json:"exif"`
	Meta       PictureMeta `json:"meta,omitempty"`
	RootPath   string      `json:"root_path,omitempty"`
}

type PictureMeta struct {
	PostedToIG   bool      `json:"posted_to_IG,omitempty"`
	Visibility   string    `json:"visibility,omitempty"`
	DateAdded    time.Time `json:"date_added,omitempty"`
	DateModified time.Time `json:"date_modified,omitempty"`
}

func (p *Picture) Save() {
	Cache.DB.Save(p)
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

func (picture *Picture) Delete() error {
	err := os.Remove(picture.Path)
	if err != nil {
		return err
	}
	Cache.DB.DeleteStruct(&picture)
	return nil
}

func (u *Picture) CreateExif() error {
	f, _ := os.Open(u.Path)
	defer f.Close()
	u.Exif = Exif{}

	m, err := imagemeta.Parse(f)
	if err != nil {
		return err
	}
	exif, err := m.Exif()
	if err != nil {
		return err
	}

	if a, err := exif.Aperture(); err == nil {
		u.Exif.FStop = float64(a)
	}
	if a, err := exif.FocalLength(); err == nil {
		u.Exif.FocalLength = float64(a)
	}
	if a, err := exif.ShutterSpeed(); err == nil {
		u.Exif.ShutterSpeed = fmt.Sprintf("%d/%d", a[0], a[1])
	}
	if a, err := exif.ISOSpeed(); err == nil {
		u.Exif.ISO = fmt.Sprint(a)
	}
	if a := exif.CameraModel(); a != "" {
		u.Exif.Camera = a
	}
	if a, err := exif.LensModel(); err == nil {
		u.Exif.LensModel = a
	}
	if a, err := exif.DateTime(nil); err == nil {
		u.Exif.DateTaken = a
	}

	lat, long, err := exif.GPSCoords()
	if err != nil {
		u.Exif.GPS = GPS{
			Lat: lat,
			Lng: long,
		}
	}
	return nil
}

func GetPictures() []Picture {
	var pics []Picture
	Cache.DB.All(&pics)
	return pics
}
func GetPictureByID(id string) (Picture, error) {
	var pic Picture
	Cache.DB.One("Id", id, &pic)
	if pic.Id == "" {
		return pic, fmt.Errorf("no picture with id %s cound be found", id)
	}
	alb, err := GetAlbumByID(pic.Album)
	if err != nil {
		return pic, err
	}
	pic.AlbumName = alb.Name
	return pic, nil
}
func GetPicturesByAlbumID(id string) []Picture {
	var pic []Picture
	Cache.DB.Find("Album", id, &pic)
	return SortByTime(pic)
}

func GetAlbumByID(id string) (Album, error) {
	var album Album
	Cache.DB.One("Id", id, &album)
	if album.Id == "" {
		return album, fmt.Errorf("no album with id %s cound be found", id)
	}
	return album, nil
}

func GetFilteredPictures(admin bool) []Picture {
	var filterPics []Picture
	for _, pic := range GetPictures() {
		if admin {
			filterPics = append(filterPics, pic)
		} else if !IsAlbumInBlacklist(pic.Album) && pic.Meta.Visibility == "PUBLIC" {
			var album Album
			Cache.DB.One("Id", pic.Album, &album)
			cleanpic := Picture{
				Id:         pic.Id,
				Name:       pic.Name,
				Caption:    pic.Caption,
				Album:      pic.Album,
				AlbumName:  album.Name,
				FormatTime: pic.Exif.DateTaken.Format("01-02-2006 15:04:05"),
				Exif:       pic.Exif,
				Meta:       pic.Meta,
				Ext:        pic.Ext,
			}
			filterPics = append(filterPics, cleanpic)
		}
	}
	return SortByTime(filterPics)
}

func SortByTime(p []Picture) []Picture {
	sort.Slice(p, func(i, j int) bool {
		return p[i].Exif.DateTaken.Sub(p[j].Exif.DateTaken) > 0
	})
	return p
}

func GetLatestPhotoDate() time.Time {
	pics := GetPictures()
	latests := pics[0].Exif.DateTaken
	for _, p := range pics {
		if p.Exif.DateTaken.After(latests) {
			latests = p.Exif.DateTaken
		}
	}
	return latests
}

func GetLatestAlbumId() string {
	pics := GetPictures()
	latests := pics[0].Exif.DateTaken
	album := pics[0].Album
	for _, p := range pics {
		if p.Exif.DateTaken.After(latests) {
			latests = p.Exif.DateTaken
			album = p.Album
		}
	}
	return album
}

func GetPhotosByDate(yourDate time.Time) []Picture {
	pics := GetPictures()
	latests := []Picture{}
	for _, p := range pics {
		if DateEqual(p.Exif.DateTaken, yourDate) {
			latests = append(latests, p)
		}
	}
	return latests
}

func IsAlbumInBlacklist(album string) bool {
	if strings.EqualFold(album, "instagram") {
		return true
	}
	if strings.EqualFold(album, "images") {
		return true
	}
	if strings.EqualFold(album, "temp") {
		return true
	}
	if strings.EqualFold(album, "rubish") {
		return true
	}
	for _, n := range config.Config.Gallery.AlbumBlacklist {
		if strings.EqualFold(album, n) {
			return true
		}
	}
	return false
}

func IsPictureInBlacklist(name string) bool {
	for _, n := range config.Config.Gallery.PictureBlacklist {
		if strings.EqualFold(name, n) {
			return true
		}
	}
	return false
}

func DoesPictureExist(p Picture) bool {
	_, err := GetPictureByID(p.Id)
	return err == nil
}
