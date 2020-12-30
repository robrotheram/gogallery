package datastore

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/araddon/dateparse"

	"github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
)

var logName = "exif"

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
	fmt.Printf("Parsing Loc: %s \n", x)
	if len(x) == 0 || len(ref) == 0 {
		return float64(0)
	}
	x = strings.Replace(x, ".", "", -1)
	parse := strings.TrimSuffix(trimFirstRune(x), "]")
	locStr := strings.Split(parse, ",")

	loc := float64(0)
	if len(locStr) > 0 {
		degStr := strings.Split(locStr[0], "/")
		deg := strToInt(degStr[0]) / strToInt(degStr[1])
		loc = float64(deg)
	}
	if len(locStr) > 1 {
		minsStr := strings.Split(locStr[1], "/")
		mins := strToInt(minsStr[0]) / strToInt(minsStr[1])
		loc = loc + float64(mins)/60
	}
	if len(locStr) > 2 {
		secStr := strings.Split(locStr[2], "/")
		secs := strToInt(secStr[0]) / strToInt(secStr[1])
		loc = loc + float64(secs)/3600
	}
	//If the location ref is either South or West then the location needs to negative.
	if ref == "S" || ref == "W" {
		return -loc
	}
	return loc
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
