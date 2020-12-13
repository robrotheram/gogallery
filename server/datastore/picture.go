package datastore

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

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

func trimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}

func strToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func convertExifGPSToFloat(x string, ref string) float64 {
	parse := strings.TrimSuffix(trimFirstRune(x), "]")
	locStr := strings.Split(parse, ",")

	degStr := strings.Split(locStr[0], "/")
	minsStr := strings.Split(locStr[1], "/")
	secStr := strings.Split(locStr[2], "/")

	deg := strToInt(degStr[0]) / strToInt(degStr[1])
	mins := strToInt(minsStr[0]) / strToInt(minsStr[1])
	secs := strToInt(secStr[0]) / strToInt(secStr[1])

	loc := float64(deg) + float64(mins)/60 + float64(secs)/3600

	//If the location ref is either South or West then the location needs to negative.
	if ref == "S" || ref == "W" {
		return -loc
	}
	return loc
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
				convertTime(decodeExifTag(x, exif.DateTime)),
				GPS{},
			}
			lat := convertExifGPSToFloat(decodeExifTag(x, exif.GPSLatitude), decodeExifTag(x, exif.GPSLatitudeRef))
			lng := convertExifGPSToFloat(decodeExifTag(x, exif.GPSLongitude), decodeExifTag(x, exif.GPSLongitudeRef))
			if lat != 0 && lng != 0 {
				u.Exif.GPS = GPS{
					Lat: lat,
					Lng: lng,
				}
			}

		}
	}
}
