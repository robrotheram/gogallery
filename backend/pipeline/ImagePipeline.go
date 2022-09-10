package pipeline

import (
	"os"
	"path/filepath"

	"github.com/h2non/bimg"
	"github.com/robrotheram/gogallery/backend/datastore/models"
	templateengine "github.com/robrotheram/gogallery/backend/templateEngine"
)

func ProcessImage(pic models.Picture) []byte {
	buffer, _ := bimg.Read(pic.Path)
	newImage, _ := bimg.NewImage(buffer).Resize(templateengine.ImageSizes["xsmall"], 0)
	return newImage
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

	buffer, err := bimg.Read(pic.Path)
	if err != nil {
		return err
	}

	for key, size := range toRender {
		cachePath := filepath.Join(destPath, key+".webp")
		newImage, _ := bimg.NewImage(buffer).Resize(size, 0)
		bimg.Write(cachePath, newImage)
	}

	return nil
}
