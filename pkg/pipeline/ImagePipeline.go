package pipeline

import (
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"image"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"

	"github.com/bep/gowebp/libwebp"
	"github.com/bep/gowebp/libwebp/webpoptions"
	"github.com/disintegration/imaging"
)

func ProcessImage(src image.Image, size int, encodeType config.ImageType, w io.Writer) {
	// Use a faster filter for small images, Lanczos for larger ones
	var filter imaging.ResampleFilter
	switch {
	case size <= 350:
		filter = imaging.Linear // fast, good for thumbnails
	case size <= 640:
		filter = imaging.CatmullRom // good quality, faster than Lanczos
	default:
		filter = imaging.Lanczos // best for large images
	}
	if size > 0 {
		src = imaging.Resize(src, size, 0, filter)
	}

	switch encodeType {
	case config.JPEG:
		jpeg.Encode(w, src, &jpeg.Options{
			Quality: 85,
		})
	case config.WebP:
		libwebp.Encode(w, src, webpoptions.EncodingOptions{
			Quality:        85,
			EncodingPreset: webpoptions.EncodingPresetPhoto,
			UseSharpYuv:    size == 0, // better color and performance
		})
	}
}
func ImageGenV2(pic datastore.Picture) error {
	destPath := filepath.Join(imgDir, pic.Id)
	os.MkdirAll(destPath, os.ModePerm)

	toRender := map[string]int{}
	for key, size := range config.ImageSizes {
		cachePath := filepath.Join(destPath, key+".webp")
		if !config.FileExists(cachePath) {
			toRender[key] = size.ImgWidth
		}
	}

	if config.Config.Gallery.UseOriginal {
		orginalPath := filepath.Join(destPath, "original"+pic.Ext)
		if !config.FileExists(orginalPath) {
			config.Copy(pic.Path, orginalPath)
		}
	}

	if len(toRender) == 0 {
		return nil
	}

	src, err := pic.Load()
	if err != nil {
		return err
	}

	for key, size := range toRender {
		cachePath := filepath.Join(destPath, key+".webp")
		fo, err := os.Create(cachePath)
		if err != nil {
			continue
		}
		defer fo.Close()
		ProcessImage(src, size, config.WebP, fo)
	}
	return nil
}
