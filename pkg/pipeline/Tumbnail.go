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
			return fmt.Errorf("load thumbnail source %s: %w", pic.Path, err)
		}
		var encoded bytes.Buffer
		if err := ProcessImage(src, config.ImageSizes[size].ImgWidth, config.JPEG, &encoded); err != nil {
			return fmt.Errorf("encode thumbnail %s: %w", pic.Path, err)
		}
		cache, err := r.ImageCache.Writer(pic.Id, config.JPEG, size)
		if err != nil {
			return fmt.Errorf("create cache writer for %s: %w", pic.Path, err)
		}
		if _, err := cache.Write(encoded.Bytes()); err != nil {
			_ = cache.Abort()
			return fmt.Errorf("write thumbnail %s: %w", pic.Path, err)
		}
		if err := cache.Close(); err != nil {
			return fmt.Errorf("save thumbnail %s: %w", pic.Path, err)
		}
		return nil
	}
}
