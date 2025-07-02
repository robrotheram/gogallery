package pipeline

import (
	"fmt"
	"testingFyne/pkg/config"
	"testingFyne/pkg/datastore"
)

func (r *RenderPipeline) generateThumbnails() func(pic datastore.Picture) error {
	return func(pic datastore.Picture) error {
		size := "small" // Default size
		if _, err := r.ImageCache.Get(pic.Id, config.JPEG, size); err == nil {
			return nil // Skip if thumbnail already exists
		}
		src, err := pic.Load()
		if err != nil {
			fmt.Printf("Failed to load image %s: %v\n", pic.Id, err)
			return fmt.Errorf("failed to load image %s: %w", pic.Id, err)
		}
		cache, err := r.ImageCache.Writer(pic.Id, config.JPEG, size)
		if err != nil {
			fmt.Printf("Failed to get cache writer for %s: %v\n", pic.Id, err)
			return fmt.Errorf("failed to get cache writer for %s: %w", pic.Id, err)

		}
		ProcessImage(src, config.ImageSizes[size].ImgWidth, config.JPEG, cache)
		return nil
	}
}
