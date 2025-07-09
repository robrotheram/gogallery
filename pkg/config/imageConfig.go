package config

type ImgSize struct {
	MinWidth int // Minimum screen width in pixels for this image source
	ImgWidth int // Recommended image width to generate for this breakpoint
}

var ImageSizes = map[string]ImgSize{
	"xsmall": {MinWidth: 0, ImgWidth: 200},     // Phones (default)
	"small":  {MinWidth: 480, ImgWidth: 400},   // Small tablets / landscape phones
	"medium": {MinWidth: 768, ImgWidth: 960},   // Tablets
	"large":  {MinWidth: 1024, ImgWidth: 1280}, // Laptops / small desktops
	"xlarge": {MinWidth: 1440, ImgWidth: 0},    // Large desktops (0 means use original size)
}

type ImageType int

const (
	JPEG ImageType = iota
	WebP
)
