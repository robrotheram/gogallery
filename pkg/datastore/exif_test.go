package datastore

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateExifKeepsFallbackMetadataForImageWithoutExif(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plain.png")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	source := image.NewRGBA(image.Rect(0, 0, 8, 4))
	source.Set(0, 0, color.White)
	if err := png.Encode(file, source); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	modified := time.Now().Add(-time.Hour)
	picture := createPicture("plain", path, "album", modified)
	_ = picture.CreateExif()
	if picture.Dimension != "8x4" || picture.AspectRatio != 2 {
		t.Fatalf("image dimensions = %q (%.2f), want 8x4 (2.0)", picture.Dimension, picture.AspectRatio)
	}
	if !picture.DateTaken.Equal(modified) {
		t.Fatalf("fallback date = %v, want %v", picture.DateTaken, modified)
	}
}

func TestApertureToStringRejectsMalformedValue(t *testing.T) {
	if got := apertureToString(map[string]string{"FNumber": "bad/value"}); got != "0.0" {
		t.Fatalf("apertureToString() = %q, want 0.0", got)
	}
}
