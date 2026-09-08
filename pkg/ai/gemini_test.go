package ai

import (
	"reflect"
	"testing"
)

func TestNormalizeMetadata(t *testing.T) {
	metadata := &ImageMetadata{
		Title:       "  Light Across the Valley  ",
		Description: "Morning light\n  crosses a quiet mountain valley.",
		Tags:        []string{" #Landscape ", "landscape", "GOLDEN   HOUR", "", "##Mountains"},
	}

	if err := normalizeMetadata(metadata); err != nil {
		t.Fatalf("normalizeMetadata() error = %v", err)
	}

	if metadata.Title != "Light Across the Valley" {
		t.Errorf("Title = %q", metadata.Title)
	}
	if metadata.Description != "Morning light crosses a quiet mountain valley." {
		t.Errorf("Description = %q", metadata.Description)
	}
	wantTags := []string{"landscape", "golden hour", "mountains"}
	if !reflect.DeepEqual(metadata.Tags, wantTags) {
		t.Errorf("Tags = %#v, want %#v", metadata.Tags, wantTags)
	}
}

func TestNormalizeMetadataRejectsIncompleteResponse(t *testing.T) {
	metadata := &ImageMetadata{Title: "A title", Tags: []string{"nature"}}
	if err := normalizeMetadata(metadata); err == nil {
		t.Fatal("normalizeMetadata() returned nil for an incomplete response")
	}
}

func TestSupportedImageMIME(t *testing.T) {
	for _, mimeType := range []string{"image/jpeg", "image/png", "image/gif", "image/webp"} {
		if !supportedImageMIME(mimeType) {
			t.Errorf("expected %q to be supported", mimeType)
		}
	}
	if supportedImageMIME("application/octet-stream") {
		t.Fatal("expected non-image content to be rejected")
	}
}
