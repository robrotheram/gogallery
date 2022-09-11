package models

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/evanoberholster/imagemeta"
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

func SortByTime(p []Picture) []Picture {
	sort.Slice(p, func(i, j int) bool {
		return p[i].Exif.DateTaken.Sub(p[j].Exif.DateTaken) > 0
	})
	return p
}
