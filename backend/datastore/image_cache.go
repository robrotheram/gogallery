package datastore

import (
	"fmt"
	"os"
	"path"
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

func (ic *ImageCache) Get(name string, size string) (*os.File, error) {
	file, err := os.Open(path.Join(ic.base, fmt.Sprintf("%s-%s.webp", name, size)))
	if err != nil {
		return nil, fmt.Errorf("file not found: %s-%s.webp", name, size)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("could not retrieve file info: %v", err)
	}

	if stat.Size() == 0 {
		file.Close()
		return nil, fmt.Errorf("file is empty: %s-%s.webp", name, size)
	}

	return file, nil
}
func (ic *ImageCache) Writer(name string, size string) (*os.File, error) {
	return os.Create(path.Join(ic.base, fmt.Sprintf("%s-%s.webp", name, size)))
}
