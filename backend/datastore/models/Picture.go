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

	meta, err := imagemeta.Decode(f)
	if err != nil {
		return err
	}

	u.Exif.FStop = meta.FNumber.String()
	u.Exif.FocalLength = meta.FocalLength.String()
	u.Exif.ShutterSpeed = meta.ExposureTime.String()
	u.Exif.ISO = fmt.Sprintf("%d", meta.ISOSpeed)
	u.Exif.Dimension = fmt.Sprintf("%d/%d", meta.ImageWidth, meta.ImageHeight)
	u.Exif.Camera = meta.CameraMake.String()
	u.Exif.LensModel = meta.LensModel
	u.Exif.DateTaken = meta.CreateDate()
	u.Exif.GPS = GPS{
		Lat: meta.GPS.Latitude(),
		Lng: meta.GPS.Latitude(),
	}

	return nil
}

func SortByTime(p []Picture) []Picture {
	sort.Slice(p, func(i, j int) bool {
		return p[i].Exif.DateTaken.Sub(p[j].Exif.DateTaken) > 0
	})
	return p
}
