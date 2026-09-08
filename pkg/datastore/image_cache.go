package datastore

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gogallery/pkg/config"
)

type ImageCache struct {
	base string
}

type CacheWriter interface {
	io.WriteCloser
	Abort() error
}

func NewImageCache() (*ImageCache, error) {
	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("locate user cache: %w", err)
	}
	cachePath := filepath.Join(cacheRoot, "gogallery", "thumbnails")
	if err := os.MkdirAll(cachePath, 0o700); err != nil {
		return nil, fmt.Errorf("create image cache: %w", err)
	}
	if err := os.Chmod(cachePath, 0o700); err != nil { // #nosec G302 -- private directory requires execute permission
		return nil, fmt.Errorf("secure image cache: %w", err)
	}
	return &ImageCache{base: cachePath}, nil
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
	cachePath, err := ic.cachePath(name, encodeType, size)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(cachePath) // #nosec G304 -- both cache-key components are strictly validated above
	if err != nil {
		return nil, fmt.Errorf("file not found: %s-%s.%s", name, size, extension(encodeType))
	}

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("could not retrieve file info: %v", err)
	}

	if stat.Size() == 0 {
		_ = file.Close()
		return nil, fmt.Errorf("file is empty: %s-%s.%s", name, size, extension(encodeType))
	}

	return file, nil
}

// Writer writes to a temporary file and publishes the completed thumbnail on
// Close. Readers therefore never observe a half-written image while background
// generation and the visible grid are working at the same time.
func (ic *ImageCache) Writer(name string, encodeType config.ImageType, size string) (CacheWriter, error) {
	target, err := ic.cachePath(name, encodeType, size)
	if err != nil {
		return nil, err
	}
	file, err := os.CreateTemp(ic.base, ".thumbnail-*")
	if err != nil {
		return nil, err
	}
	return &atomicCacheWriter{File: file, target: target}, nil
}

func (ic *ImageCache) cachePath(name string, encodeType config.ImageType, size string) (string, error) {
	if !validCacheComponent(name) || !validCacheComponent(size) {
		return "", fmt.Errorf("invalid image cache key")
	}
	return filepath.Join(ic.base, fmt.Sprintf("%s-%s.%s", name, size, extension(encodeType))), nil
}

func validCacheComponent(value string) bool {
	if value == "" || len(value) > 128 || value == "." || value == ".." {
		return false
	}
	return strings.IndexFunc(value, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_')
	}) == -1
}

func (ic *ImageCache) Reset() {
	if err := os.RemoveAll(ic.base); err != nil {
		fmt.Printf("Failed to reset image cache: %v\n", err)
	}
	if err := os.MkdirAll(ic.base, 0o700); err != nil {
		fmt.Printf("Failed to recreate image cache directory: %v\n", err)
	}
}

type atomicCacheWriter struct {
	*os.File
	target string
	closed bool
}

func (w *atomicCacheWriter) Abort() error {
	if w.closed {
		return nil
	}
	w.closed = true
	temporaryPath := w.File.Name()
	closeErr := w.File.Close()
	removeErr := os.Remove(temporaryPath)
	if closeErr != nil {
		return closeErr
	}
	if removeErr != nil && !os.IsNotExist(removeErr) {
		return removeErr
	}
	return nil
}

func (w *atomicCacheWriter) Close() error {
	if w.closed {
		return nil
	}
	w.closed = true
	tempName := w.File.Name()
	if err := w.File.Close(); err != nil {
		_ = os.Remove(tempName)
		return err
	}
	if err := os.Rename(tempName, w.target); err != nil {
		_ = os.Remove(tempName)
		return err
	}
	return nil
}
