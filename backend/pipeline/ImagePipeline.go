package pipeline

import (
	"io"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/robrotheram/gogallery/backend/datastore/models"
	templateengine "github.com/robrotheram/gogallery/backend/templateEngine"
)

func ProcessImage(pic models.Picture, w io.Writer) {
	srcImage, err := imaging.Open(pic.Path)
	if err != nil {
		log.Fatalf("failed to open image: %v", err)
	}

	size := templateengine.ImageSizes["xsmall"]
	newImage := imaging.Resize(srcImage, size, size, imaging.Lanczos)
	fileFormat, formatErr := imaging.FormatFromFilename(pic.Path)
	if formatErr != nil {
		log.Fatalf("failed to detect image format: %v", err)
	}
	imaging.Encode(w, newImage, fileFormat)
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

	for key, size := range toRender {
		cachePath := filepath.Join(destPath, key+".webp")
		oldImage, err := imaging.Open(pic.Path)
		if err != nil {
			return err
		}

		newImage := imaging.Resize(oldImage, size, size, imaging.Lanczos)
		err = imaging.Save(newImage, cachePath)
		if err != nil {
			return err
		}
	}

	return nil
}
