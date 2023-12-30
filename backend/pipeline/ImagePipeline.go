package pipeline

import (
	"image"
	"io"
	"os"
	"path/filepath"

	"github.com/bep/gowebp/libwebp"
	"github.com/bep/gowebp/libwebp/webpoptions"
	"github.com/disintegration/imaging"
	"github.com/robrotheram/gogallery/backend/datastore/models"
	templateengine "github.com/robrotheram/gogallery/backend/templateEngine"
)

func resize(base image.Image, width int, height int) image.Image {
	if width == 0 && height == 0 {
		return imaging.Resize(base, int(float64(base.Bounds().Dx())), 0, imaging.Lanczos)
	}
	return imaging.Resize(base, width, height, imaging.Lanczos)
}

func ProcessImage(src image.Image, size int, w io.Writer) {
	resized := resize(src, size, 0)
	libwebp.Encode(w, resized, webpoptions.EncodingOptions{
		Quality:        100,
		EncodingPreset: webpoptions.EncodingPresetDefault,
		UseSharpYuv:    true,
	})
}

func ImageGenV2(pic models.Picture) error {
	destPath := filepath.Join(imgDir, pic.Id)
	os.MkdirAll(destPath, os.ModePerm)

	toRender := map[string]int{}
	for key, size := range templateengine.ImageSizes {
		cachePath := filepath.Join(destPath, key+".webp")
		if !templateengine.FileExists(cachePath) {
			toRender[key] = size
		}
	}

	orginalPath := filepath.Join(destPath, "original"+pic.Ext)
	if !templateengine.FileExists(orginalPath) {
		templateengine.File(pic.Path, orginalPath)
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
