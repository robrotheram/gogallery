package pipeline

import (
	"bytes"
	"fmt"
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
)

func (r *RenderPipeline) generateThumbnails() func(pic datastore.Picture) error {
	return func(pic datastore.Picture) error {
		size := "small" // Default size
		if cached, err := r.ImageCache.Get(pic.Id, config.JPEG, size); err == nil {
			_ = cached.Close()
			return nil // Skip if thumbnail already exists
		}
		src, err := pic.Load()
		if err != nil {
			fmt.Printf("Failed to load image %s: %v\n", pic.Id, err)
			return fmt.Errorf("failed to load image %s: %w", pic.Id, err)
		}
		var encoded bytes.Buffer
		if err := ProcessImage(src, config.ImageSizes[size].ImgWidth, config.JPEG, &encoded); err != nil {
			return fmt.Errorf("encode thumbnail %s: %w", pic.Id, err)
		}
		cache, err := r.ImageCache.Writer(pic.Id, config.JPEG, size)
		if err != nil {
			return fmt.Errorf("create cache writer for %s: %w", pic.Id, err)
		}
		if _, err := cache.Write(encoded.Bytes()); err != nil {
			_ = cache.Abort()
			return fmt.Errorf("write thumbnail %s: %w", pic.Id, err)
		}
		if err := cache.Close(); err != nil {
			return fmt.Errorf("failed to save thumbnail %s: %w", pic.Id, err)
		}
		return nil
	}
}
