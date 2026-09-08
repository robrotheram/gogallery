package pipeline

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"

	"gogallery/pkg/config"
	"gogallery/pkg/datastore"

	"github.com/bep/gowebp/libwebp"
	"github.com/bep/gowebp/libwebp/webpoptions"
	"github.com/disintegration/imaging"
)

func ProcessImage(src image.Image, size int, encodeType config.ImageType, w io.Writer) error {
	if src == nil {
		return fmt.Errorf("source image is nil")
	}
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
		return jpeg.Encode(w, src, &jpeg.Options{
			Quality: 85,
		})
	case config.WebP:
		return libwebp.Encode(w, src, webpoptions.EncodingOptions{
			Quality:        85,
			EncodingPreset: webpoptions.EncodingPresetPhoto,
			UseSharpYuv:    size == 0, // better color and performance
		})
	default:
		return fmt.Errorf("unsupported image encoding %d", encodeType)
	}
}

func (r *RenderPipeline) imageGenV2(pic datastore.Picture) error {
	destPath := filepath.Join(r.imgDir, pic.Id)
	// #nosec G301 -- generated website directories must be readable by a web server.
	if err := os.MkdirAll(destPath, 0o755); err != nil {
		return err
	}

	toRender := map[string]int{}
	for key, size := range config.ImageSizes {
		cachePath := filepath.Join(destPath, key+".webp")
		if !config.FileExists(cachePath) {
			toRender[key] = size.ImgWidth
		}
	}

	if r.config.UseOriginal {
		originalPath := filepath.Join(destPath, "original"+pic.Ext)
		if !config.FileExists(originalPath) {
			if err := config.Copy(pic.Path, originalPath); err != nil {
				return err
			}
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
		if err := writeGeneratedFile(cachePath, func(w io.Writer) error {
			return ProcessImage(src, size, config.WebP, w)
		}); err != nil {
			return fmt.Errorf("write image size %s: %w", key, err)
		}
	}
	return nil
}
