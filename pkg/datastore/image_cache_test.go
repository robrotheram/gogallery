package datastore

import (
	"gogallery/pkg/config"
	"io"
	"strings"
	"testing"
)

func TestImageCachePublishesCompletedWritesAtomically(t *testing.T) {
	cache := &ImageCache{base: t.TempDir()}
	w, err := cache.Writer("photo", config.JPEG, "small")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("complete thumbnail")); err != nil {
		t.Fatal(err)
	}

	if file, err := cache.Get("photo", config.JPEG, "small"); err == nil {
		file.Close()
		t.Fatal("cache entry became visible before the writer was closed")
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	file, err := cache.Get("photo", config.JPEG, "small")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), "complete thumbnail"; got != want {
		t.Fatalf("cache data = %q, want %q", got, want)
	}
}

func TestImageCacheRejectsPathTraversal(t *testing.T) {
	cache := &ImageCache{base: t.TempDir()}
	for _, value := range []string{"../outside", "small/../../outside", "", "."} {
		if _, err := cache.Writer("photo", config.JPEG, value); err == nil || !strings.Contains(err.Error(), "invalid image cache key") {
			t.Fatalf("Writer accepted unsafe cache size %q: %v", value, err)
		}
	}
}

func TestImageCacheAbortDoesNotPublishPartialWrite(t *testing.T) {
	cache := &ImageCache{base: t.TempDir()}
	writer, err := cache.Writer("photo", config.JPEG, "small")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("partial")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Abort(); err != nil {
		t.Fatal(err)
	}
	if _, err := cache.Get("photo", config.JPEG, "small"); err == nil {
		t.Fatal("aborted cache entry was published")
	}
}
