package datastore

import (
	"fmt"
	"image"
	"io"
	"os"
	"slices"
	"time"

	"github.com/bep/imagemeta"
)

type Exif struct {
	FStop        string    `json:"f_stop"`
	FocalLength  string    `json:"focal_length"`
	ShutterSpeed string    `json:"shutter_speed"`
	ISO          string    `json:"iso"`
	Dimension    string    `json:"dimension"`
	AspectRatio  float32   `json:"aspect_ratio"`
	Camera       string    `json:"camera"`
	LensModel    string    `json:"lens_model"`
	DateTaken    time.Time `json:"date_taken"`
	GPS          GPS       `json:"gps"`
	FileFormat   string    `json:"file_format"`
	Software     string    `json:"software"`
	ColorSpace   string    `json:"color_space"`
	FocusMode    string    `json:"focus_mode"`
	MeteringMode string    `json:"metering_mode"`

	WhiteBalance string `json:"white_balance,omitempty"`
	Saturation   string `json:"saturation,omitempty"`
	Contrast     string `json:"contrast,omitempty"`
	Sharpness    string `json:"sharpness,omitempty"`
	Temperature  string `json:"temperature,omitempty"`
	Cropped      string `json:"cropped,omitempty"`
}

type GPS struct {
	Lat float64 `json:"latitude"`
	Lng float64 `json:"longitude"`
}

// Change CreateExif to take a pointer to Picture so it can modify the struct in-place
func CreateExif(u *Picture) error {
	f, _ := os.Open(u.Path)
	defer f.Close()

	var tags imagemeta.Tags
	handleTag := func(ti imagemeta.TagInfo) error {
		tags.Add(ti)
		return nil
	}

	err := imagemeta.Decode(imagemeta.Options{R: f, ImageFormat: imagemeta.JPEG, HandleTag: handleTag})
	if err != nil {
		return err
	}
	meta := tags.All()

	if latitude, longitude, err := tags.GetLatLong(); err != nil {
		u.GPSLat = latitude
		u.GPSLng = longitude
	}
	if found, err := tags.GetDateTime(); err != nil {
		u.DateTaken = found
	} else {
		if meta["CreateDate"].Value != nil {
			timeStr := meta["CreateDate"].Value.(string)
			if parsedTime, err := ParseExifDateTime(timeStr); err == nil {
				u.DateTaken = parsedTime
			}
		}
	}

	if w, h, err := GetImageDention(f); err == nil {
		u.Dimension = fmt.Sprintf("%dx%d", w, h)
		u.AspectRatio = float32(w) / float32(h)
	}

	u.FStop = fmt.Sprint(meta["FNumber"].Value)
	u.FocalLength = fmt.Sprint(meta["FocalLength"].Value)
	u.ShutterSpeed = fmt.Sprint(meta["ExposureTime"].Value)
	u.ISO = fmt.Sprint(meta["ISO"].Value)
	u.Camera = FormatCamera(meta)
	if u.Camera == "<nil> <nil>" {
		u.Camera = ""
	}
	u.LensModel = fmt.Sprint(meta["LensModel"].Value)

	u.FileFormat = fmt.Sprint(meta["ImageType"].Value)
	u.Software = fmt.Sprint(meta["Software"].Value)

	u.ColorSpace = ColorSpaceToString(meta["ColorSpace"].Value)
	u.MeteringMode = MeteringModeToString(meta["MeteringMode"].Value)

	u.Saturation = fmt.Sprint(meta["Saturation"].Value)
	u.Contrast = fmt.Sprint(meta["Contrast"].Value)
	u.Sharpness = fmt.Sprint(meta["Sharpness"].Value)
	u.Temperature = fmt.Sprint(meta["Temperature"].Value)
	u.WhiteBalance = fmt.Sprint(meta["WhiteBalance"].Value)
	u.Cropped = CroppedToString(meta["HasCrop"].Value)

	CleanExif(u)
	return nil
}
func GetImageDention(f *os.File) (width, height int, err error) {
	f.Seek(0, io.SeekStart) // Reset the reader to the beginning
	img, format, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, fmt.Errorf("image decode config failed: %v", err)
	}
	if format != "jpeg" && format != "png" && format != "gif" && format != "webp" {
		return 0, 0, fmt.Errorf("unsupported image format: %s", format)
	}
	return img.Width, img.Height, nil
}

func FormatCamera(tags map[string]imagemeta.TagInfo) string {
	make, _ := tags["Make"].Value.(string)
	model, _ := tags["Model"].Value.(string)

	if make == "" && model == "" {
		return ""
	}

	return fmt.Sprintf("%s %s", make, model)
}

func CroppedToString(val any) string {
	hasCropped, _ := val.(string)
	if hasCropped == "True" {
		return "Yes"
	}
	return "No"
}
func ColorSpaceToString(val any) string {
	cs, _ := val.(uint16)
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

func MeteringModeToString(val any) string {
	mm, _ := val.(uint16)
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

func CleanExif(p *Picture) {
	fields := map[*string][]string{
		&p.FStop:        {"<nil>", "0"},
		&p.FocalLength:  {"<nil>", "0"},
		&p.ShutterSpeed: {"<nil>", "0"},
		&p.ISO:          {"<nil>", "0"},
		&p.Dimension:    {"<nil>", "0x0"},
		&p.Camera:       {"<nil>", "0"},
		&p.LensModel:    {"<nil>", "0"},
		&p.FileFormat:   {"<nil>", "0"},
		&p.Software:     {"<nil>", "0"},
		&p.ColorSpace:   {"<nil>", "0"},
		&p.MeteringMode: {"<nil>", "0"},
		&p.Saturation:   {"<nil>", "0"},
		&p.Contrast:     {"<nil>", "0"},
		&p.Sharpness:    {"<nil>", "0"},
		&p.Temperature:  {"<nil>", "0"},
		&p.WhiteBalance: {"<nil>", "0"},
		&p.Cropped:      {"<nil>", "0"},
	}

	for field, invalidValues := range fields {
		if slices.Contains(invalidValues, *field) {
			*field = ""
		}
	}
}

// ParseExifDateTime tries to parse EXIF date/time strings in various common formats
func ParseExifDateTime(s string) (time.Time, error) {
	layouts := []string{
		"2006:01:02 15:04:05", // EXIF standard
		"2006-01-02T15:04:05", // ISO8601 without timezone
		time.RFC3339,          // ISO8601 with timezone
		time.RFC822,           // RFC822
		time.RFC1123,          // RFC1123
		time.RFC1123Z,         // RFC1123 with numeric TZ
		"2006-01-02 15:04:05", // MySQL/SQLite common
	}
	for _, layout := range layouts {
		if layout == "time.RFC3339" {
			if t, err := time.Parse(time.RFC3339, s); err == nil {
				return t, nil
			}
			continue
		}
		if layout == "time.RFC822" {
			if t, err := time.Parse(time.RFC822, s); err == nil {
				return t, nil
			}
			continue
		}
		if layout == "time.RFC1123" {
			if t, err := time.Parse(time.RFC1123, s); err == nil {
				return t, nil
			}
			continue
		}
		if layout == "time.RFC1123Z" {
			if t, err := time.Parse(time.RFC1123Z, s); err == nil {
				return t, nil
			}
			continue
		}
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse EXIF date: %q", s)
}
