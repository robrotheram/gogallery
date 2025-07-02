package datastore

import (
	"fmt"
	"image"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	jpegstructure "github.com/dsoprea/go-jpeg-image-structure/v2"
)

var exifIfdMapping *exifcommon.IfdMapping
var exifTagIndex = exif.NewTagIndex()
var exifDateTimeTags = []string{"DateTimeOriginal", "DateTimeCreated", "CreateDate", "DateTime", "DateTimeDigitized"}

func parser(fileName string) (rawExif []byte, err error) {
	jpegMp := jpegstructure.NewJpegMediaParser()
	sl, err := jpegMp.ParseFile(fileName)

	if err != nil {
		return nil, fmt.Errorf("failed to parse jpeg file %s: %w", fileName, err)
	}
	_, rawExif, err = sl.Exif()
	if err == nil {
		return rawExif, nil
	}
	rawExif, err = exif.SearchFileAndExtractExif(fileName)
	if err != nil {
		return rawExif, fmt.Errorf("found no exif data")
	}
	return rawExif, nil
}

func (u *Picture) parseGPS(rawExif []byte) error {
	var ifdIndex exif.IfdIndex
	_, ifdIndex, err := exif.Collect(exifIfdMapping, exifTagIndex, rawExif)

	if err != nil {
		return err
	}

	return u.extractGPSData(ifdIndex)
}

func (u *Picture) extractGPSData(ifdIndex exif.IfdIndex) error {
	ifd, err := ifdIndex.RootIfd.ChildWithIfdPath(exifcommon.IfdGpsInfoStandardIfdIdentity)
	if err != nil {
		return nil // No GPS data found, not an error.
	}

	gi, err := ifd.GpsInfo()
	if err != nil {
		return err
	}

	if !math.IsNaN(gi.Latitude.Decimal()) && !math.IsNaN(gi.Longitude.Decimal()) {
		u.GPSLat, u.GPSLng = NormalizeGPS(gi.Latitude.Decimal(), gi.Longitude.Decimal())
	}

	if gi.Altitude != 0 {
		u.GPSAltitude = float64(gi.Altitude)
	}

	if !gi.Timestamp.IsZero() {
		u.GPSTimestamp = gi.Timestamp
	}

	return nil
}

func (u *Picture) CreateExif() error {

	rawExif, err := parser(u.Path)
	if err != nil {
		return err
	}

	opt := exif.ScanOptions{}
	entries, _, err := exif.GetFlatExifData(rawExif, &opt)
	if err != nil {
		log.Printf("Error getting flat EXIF data for %s: %v", u.Path, err)
		return err
	}

	tags := make(map[string]string, len(entries))

	// Ignore IFD1 tags with existing IFD0 values.
	for _, tag := range entries {
		s := strings.Split(tag.FormattedFirst, "\x00")
		if tag.TagName == "" || len(s) == 0 {
			// Do nothing.
		} else if s[0] != "" && (tags[tag.TagName] == "" || tag.IfdPath != exif.ThumbnailFqIfdPath) {
			tags[tag.TagName] = s[0]
		}
	}

	// Abort if no values were found.
	if len(tags) == 0 {
		return fmt.Errorf("metadata: no exif data in %s", u.Path)
	}

	u.parseGPS(rawExif)
	u.DateTaken = parseExifDateTime(tags)
	u.Camera = cameraModelToString(tags)
	u.FStop = apatureToString(tags)
	u.FocalLength = focalLengthToString(tags)

	if value, ok := tags["FocalLengthIn35mmFilm"]; ok {
		u.FocalLength = value
	} else {
		u.FocalLength = tags["FocalLength"]
	}

	u.ISO = tags["ISOSpeedRatings"]
	u.ShutterSpeed = tags["ExposureTime"]
	u.LensModel = tags["LensModel"]
	u.FileFormat = tags["FileType"]
	u.Software = tags["Software"]

	if w, h, err := GetImageDention(u); err == nil {
		u.Dimension = fmt.Sprintf("%dx%d", w, h)
		u.AspectRatio = float32(w) / float32(h)
	}

	u.ColorSpace = formatColorSpace(tags["ColorSpace"])
	u.MeteringMode = formatMeteringMode(tags["MeteringMode"])
	u.Saturation = formatSaturation(tags["Saturation"])
	u.Contrast = formatContrast(tags["Contrast"])
	u.Sharpness = formatSharpness(tags["Sharpness"])
	u.Temperature = formatTemperature(tags["Temperature"])
	u.WhiteBalance = formatWhiteBalance(tags["WhiteBalance"])

	return nil
}
func GetImageDention(u *Picture) (width, height int, err error) {
	f, err := os.Open(u.Path)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to open image file %s: %v", u.Path, err)
	}
	defer f.Close()

	img, format, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, fmt.Errorf("image decode config failed: %v", err)
	}
	if format != "jpeg" && format != "png" && format != "gif" && format != "webp" {
		return 0, 0, fmt.Errorf("unsupported image format: %s", format)
	}
	return img.Width, img.Height, nil
}

func parseExifDateTime(tags map[string]string) time.Time {
	takenAt := time.Time{}
	for _, name := range exifDateTimeTags {
		if dateTime, _ := dateparse.ParseAny(tags[name]); !dateTime.IsZero() {
			takenAt = dateTime
			break
		}
	}
	if !takenAt.IsZero() {
		if takenAtLocal, err := time.ParseInLocation("2006-01-02T15:04:05", takenAt.Format("2006-01-02T15:04:05"), time.UTC); err == nil {
			return takenAtLocal
		} else {
			return takenAt
		}
	}
	return takenAt
}

func cameraModelToString(tags map[string]string) string {
	// Extract camera make and model from EXIF tags.
	make := tags["Make"]
	model := tags["Model"]
	if make != "" && model != "" {
		return fmt.Sprintf("%s %s", make, model)
	} else if make != "" {
		return make
	} else if model != "" {
		return model
	} else {
		return "Unknown Camera"
	}
}

func apatureToString(tags map[string]string) string {
	if value, ok := tags["FNumber"]; ok {
		values := strings.Split(value, "/")
		if len(values) == 2 && values[1] != "0" && values[1] != "" {
			number, _ := strconv.ParseFloat(values[0], 64)
			denom, _ := strconv.ParseFloat(values[1], 64)
			return fmt.Sprintf("%.1f", math.Round((number/denom)*1000)/1000)
		}
	}
	return "0.0"
}

func focalLengthToString(tags map[string]string) string {
	if value, ok := tags["FocalLengthIn35mmFilm"]; ok {
		return value
	}
	return tags["FocalLength"]
}

func formatSaturation(sat string) string {
	switch sat {
	case "0", "<nil>":
		return "Normal"
	case "-1":
		return "Low"
	case "1":
		return "High"
	default:
		return "Unknown"
	}
}

func formatContrast(contrast string) string {
	switch contrast {
	case "0", "<nil>":
		return "Normal"
	case "-1":
		return "Low"
	case "1":
		return "High"
	default:
		return "Unknown"
	}
}

func formatSharpness(sharpness string) string {
	switch sharpness {
	case "0", "<nil>":
		return "Normal"
	case "-1":
		return "Low"
	case "1":
		return "High"
	default:
		return "Unknown"
	}
}

func formatTemperature(temp string) string {
	switch temp {
	case "0", "<nil>":
		return "Auto"
	case "1":
		return "Manual"
	case "2":
		return "Daylight"
	case "3":
		return "Cloudy"
	case "4":
		return "Tungsten"
	case "5":
		return "Fluorescent"
	case "6":
		return "Flash"
	default:
		return "Unknown"
	}
}

func formatWhiteBalance(wb string) string {
	switch wb {
	case "0", "<nil>":
		return "Auto"
	case "1":
		return "Manual"
	case "2":
		return "Daylight"
	case "3":
		return "Cloudy"
	case "4":
		return "Tungsten"
	case "5":
		return "Fluorescent"
	case "6":
		return "Flash"
	default:
		return "Unknown"
	}
}

func formatColorSpace(val string) string {
	cs, _ := strconv.ParseUint(val, 16, 16)
	switch cs {
	case 0x1:
		return "sRGB"
	case 0x2:
		return "Adobe RGB"
	case 0xfffd:
		return "Wide Gamut RGB"
	case 0xfffe:
		return "ICC Profile"
	case 0xffff:
		return "Uncalibrated"
	default:
		return "Unknown"
	}
}

func formatMeteringMode(val string) string {
	mm, _ := strconv.ParseUint(val, 16, 8)
	switch mm {
	case 0x1:
		return "Average"
	case 0x2:
		return "Center Weighted Average"
	case 0x3:
		return "Spot"
	case 0x4:
		return "Multi Spot"
	case 0x5:
		return "Pattern"
	case 0x6:
		return "Partial"
	case 0x7:
		return "Other"
	default:
		return "Unknown"
	}
}
