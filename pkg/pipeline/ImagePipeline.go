package pipeline

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"math"
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
	bounds := src.Bounds()
	if bounds.Empty() {
		return fmt.Errorf("source image has empty dimensions")
	}
	scale := 1.0
	if size > 0 {
		scale = float64(size) / float64(bounds.Dx())
	}
	if encodeType == config.WebP {
		// Clamp the requested output before allocating a resized image.
		// libwebp's WEBP_MAX_DIMENSION applies to both sides, so a tall
		// photo can exceed it even at one of the smaller requested widths.
		const maxDimension = 16383
		scale = min(scale, float64(maxDimension)/float64(bounds.Dx()), float64(maxDimension)/float64(bounds.Dy()))
	}
	width := max(1, int(math.Round(float64(bounds.Dx())*scale)))
	height := max(1, int(math.Round(float64(bounds.Dy())*scale)))
	// Use a faster filter for small images, Lanczos for larger ones
	var filter imaging.ResampleFilter
	switch {
	case size > 0 && size <= 350:
		filter = imaging.Linear // fast, good for thumbnails
	case size > 0 && size <= 640:
		filter = imaging.CatmullRom // good quality, faster than Lanczos
	default:
		filter = imaging.Lanczos // best for large images
	}
	if width != bounds.Dx() || height != bounds.Dy() {
		src = imaging.Resize(src, width, height, filter)
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

func (r *RenderPipeline) imageGenV2(pic datastore.Picture) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s: %w", pic.Path, err)
		}
	}()
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
