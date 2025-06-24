package pipeline

import (
	"image"
	"io"
	"os"
	"path/filepath"

	"github.com/bep/gowebp/libwebp"
	"github.com/bep/gowebp/libwebp/webpoptions"
	"github.com/disintegration/imaging"
	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/datastore"
	templateengine "github.com/robrotheram/gogallery/backend/templateEngine"
)

func ProcessImage(src image.Image, size int, w io.Writer) {
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
	resized := imaging.Resize(src, size, 0, filter)
	libwebp.Encode(w, resized, webpoptions.EncodingOptions{
		Quality:        85,
		EncodingPreset: webpoptions.EncodingPresetPhoto,
		UseSharpYuv:    true, // better color and performance
	})
}
func ImageGenV2(pic datastore.Picture) error {
	destPath := filepath.Join(imgDir, pic.Id)
	os.MkdirAll(destPath, os.ModePerm)

	toRender := map[string]int{}
	for key, size := range templateengine.ImageSizes {
		cachePath := filepath.Join(destPath, key+".webp")
		if !templateengine.FileExists(cachePath) {
			toRender[key] = size.ImgWidth
		}
	}

	if config.Config.Gallery.UseOriginal {
		orginalPath := filepath.Join(destPath, "original"+pic.Ext)
		if !templateengine.FileExists(orginalPath) {
			templateengine.Copy(pic.Path, orginalPath)
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
		ProcessImage(src, size, fo)
	}
	return nil
}
