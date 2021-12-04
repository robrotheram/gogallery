package datastore

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	"github.com/robrotheram/gogallery/config"
)

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

type PictureMeta struct {
	PostedToIG   bool      `json:"posted_to_IG,omitempty"`
	Visibility   string    `json:"visibility,omitempty"`
	DateAdded    time.Time `json: date_added,omitempty`
	DateModified time.Time `json: date_modified,omitempty`
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

func (picture *Picture) Delete() {
	err := os.Remove(picture.Path)
	if err != nil {
		fmt.Println(err)
	}
}

func (u *Picture) CreateExif() error {
	raw, err := GetRawExif(u.Path)
	if err != nil {
		return err
	}
	exifData := GetExifTags(raw)

	u.Exif = Exif{
		FStop:        fnumber(exifData["FNumber"]),
		FocalLength:  fnumber(exifData["FocalLength"]),
		ShutterSpeed: exifData["ExposureTime"],
		ISO:          exifData["ISOSpeedRatings"],
		Dimension:    fmt.Sprintf("%sx%s", exifData["PixelXDimension"], exifData["PixelYDimension"]),
		Camera:       exifData["ISOSpeedRatings"],
		LensModel:    exifData["ISOSpeedRatings"],
		DateTaken:    convertTime(exifData["DateTime"]),
		GPS:          GPS{},
	}

	var exifIfdMapping *exifcommon.IfdMapping
	var exifTagIndex = exif.NewTagIndex()

	exifIfdMapping = exifcommon.NewIfdMapping()

	if err := exifcommon.LoadStandardIfds(exifIfdMapping); err != nil {
		fmt.Printf("metadata: %s \n", err.Error())
	}

	_, index, err := exif.Collect(exifIfdMapping, exifTagIndex, raw)

	if err == nil {
		if ifd, err := index.RootIfd.ChildWithIfdPath(exifcommon.IfdGpsInfoStandardIfdIdentity); err == nil {
			if gi, err := ifd.GpsInfo(); err == nil {
				u.Exif.GPS.Lat = float64(gi.Latitude.Decimal())
				u.Exif.GPS.Lng = float64(gi.Longitude.Decimal())
				//u.Exif.GPS.Altitude = gi.Altitude
			}
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
	return pic
}

func GetAlbumByID(id string) (Album, error) {
	var album Album
	Cache.DB.One("Id", id, &album)
	if album.Id == "" {
		return album, fmt.Errorf("no album with id %s cound be found", id)
	}
	return album, nil
}
func GetFilteredPictures() []Picture {
	var filterPics []Picture
	for _, pic := range GetPictures() {
		if !IsAlbumInBlacklist(pic.Album) {
			if pic.Meta.Visibility == "PUBLIC" {
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
				}
				filterPics = append(filterPics, cleanpic)
			}
		}
	}
	sort.Slice(filterPics, func(i, j int) bool {
		return filterPics[i].Exif.DateTaken.Sub(filterPics[j].Exif.DateTaken) > 0
	})
	return filterPics
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
	for _, n := range config.Config.Gallery.AlbumBlacklist {
		if strings.EqualFold(album, n) {
			return true
		}
	}
	return false
}

func IsPictureInBlacklist(pic string) bool {
	for _, n := range config.Config.Gallery.PictureBlacklist {
		if strings.EqualFold(pic, n) {
			return true
		}
	}
	return false
}

func DoesPictureExist(p Picture) bool {
	_, err := GetPictureByID(p.Id)
	return err == nil
}
