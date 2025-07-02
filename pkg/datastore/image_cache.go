package datastore

import (
	"fmt"
	"os"
	"path"
	"testingFyne/pkg/config"
)

type ImageCache struct {
	base string
}

func NewImageCache() *ImageCache {
	path := path.Join(os.TempDir(), "gogallery")
	os.MkdirAll(path, 0755)
	return &ImageCache{
		base: path,
	}
}

func extension(encodeType config.ImageType) string {
	switch encodeType {
	case config.JPEG:
		return "jpg"
	case config.WebP:
		return "webp"
	default:
		return "jpg" // Default to JPEG if type is unknown
	}
}

func (ic *ImageCache) Get(name string, encodeType config.ImageType, size string) (*os.File, error) {

	file, err := os.Open(path.Join(ic.base, fmt.Sprintf("%s-%s.%s", name, size, extension(encodeType))))
	if err != nil {
		return nil, fmt.Errorf("file not found: %s-%s.%s", name, size, extension(encodeType))
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("could not retrieve file info: %v", err)
	}

	if stat.Size() == 0 {
		file.Close()
		return nil, fmt.Errorf("file is empty: %s-%s.%s", name, size, extension(encodeType))
	}

	return file, nil
}
func (ic *ImageCache) Writer(name string, encodeType config.ImageType, size string) (*os.File, error) {
	return os.Create(path.Join(ic.base, fmt.Sprintf("%s-%s.%s", name, size, extension(encodeType))))
}

func (ic *ImageCache) Reset() {
	if err := os.RemoveAll(ic.base); err != nil {
		fmt.Printf("Failed to reset image cache: %v\n", err)
	}
	if err := os.MkdirAll(ic.base, 0755); err != nil {
		fmt.Printf("Failed to recreate image cache directory: %v\n", err)
	}
}
