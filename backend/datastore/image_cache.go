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
	return os.Open(path.Join(ic.base, fmt.Sprintf("%s-%s.webp", name, size)))
}
func (ic *ImageCache) Writer(name string, size string) (*os.File, error) {
	return os.Create(path.Join(ic.base, fmt.Sprintf("%s-%s.webp", name, size)))
}
