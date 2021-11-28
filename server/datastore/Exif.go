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
)

var logName = "exif"

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
