package datastore

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"github.com/rwcarlsen/goexif/tiff"
)

func fnumber(f string) float64 {
	f = strings.Replace(f, "\"", "", -1)
	calc := strings.Split(f, "/")
	if len(calc) != 2 {
		return 10
	}
	if i, err := strconv.ParseFloat(calc[0], 64); err == nil {
		if j, err := strconv.ParseFloat(calc[1], 64); err == nil {
			return i / j
		}
	}
	return 11
}

func decodeExifTag(exf *exif.Exif, tag exif.FieldName) (val string) {
	res, err := exf.Get(tag)
	if err != nil {
		return ""
	}

	switch res.Format() {
	case tiff.StringVal:
		resStr, err := res.StringVal()
		if err != nil {
			fmt.Println(err)
		}
		return resStr
		break
	case tiff.RatVal:
		return strings.Replace(res.String(), "\"", "", -1)
	}
	return res.String()

}

func convertTime(str string) time.Time {
	if str == "" {
		return time.Time{}
	}
	dateTime := strings.Split(str, " ")
	date := strings.Split(dateTime[0], ":")
	t, _ := dateparse.ParseLocal(fmt.Sprintf("%s/%s/%s %s", date[0], date[1], date[2], dateTime[1]))
	return t
}

func (u *Picture) CreateExif() {
	f, err := os.Open(u.Path)
	if err == nil {
		exif.RegisterParsers(mknote.All...)
		x, err := exif.Decode(f)
		if err == nil {
			u.Exif = Exif{
				fnumber(decodeExifTag(x, exif.FNumber)),
				fnumber(decodeExifTag(x, exif.FocalLength)),
				decodeExifTag(x, exif.ExposureTime),
				decodeExifTag(x, exif.ISOSpeedRatings),
				fmt.Sprintf("%sx%s", decodeExifTag(x, exif.PixelXDimension), decodeExifTag(x, exif.PixelYDimension)),
				decodeExifTag(x, exif.Make),
				decodeExifTag(x, exif.LensModel),
				convertTime(decodeExifTag(x, exif.DateTime))}
		}
	}
}
