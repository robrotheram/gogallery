package models

import (
	"os"
	"strings"
	"time"

	"github.com/dsoprea/go-exif/v3"
)

type Exif struct {
	FStop        float64   `json:"f_stop"`
	FocalLength  float64   `json:"focal_length"`
	ShutterSpeed string    `json:"shutter_speed"`
	ISO          string    `json:"iso"`
	Dimension    string    `json:"dimension"`
	Camera       string    `json:"camera"`
	LensModel    string    `json:"lens_model"`
	DateTaken    time.Time `json:"date_taken"`
	GPS          GPS       `json:"gps"`
}

type GPS struct {
	Lat float64 `json:"latitude"`
	Lng float64 `json:"longitude"`
}

func GetRawExif(path string) ([]byte, error) {
	source, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = source.Close()
	}()
	return exif.SearchAndExtractExifWithReader(source)
}

func GetExifTags(rawExif []byte) map[string]string {

	opt := exif.ScanOptions{}
	entries, _, _ := exif.GetFlatExifData(rawExif, &opt)

	data := make(map[string]string)
	for _, entry := range entries {
		if entry.TagName != "" && entry.Formatted != "" {
			data[entry.TagName] = strings.Split(entry.FormattedFirst, "\x00")[0]
		}
	}
	return data
}
