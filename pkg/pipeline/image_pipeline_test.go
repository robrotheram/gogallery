package pipeline

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"gogallery/pkg/monitor"

	"golang.org/x/image/webp"
)

func TestImageBatchReportsEveryFailedSource(t *testing.T) {
	previousSizes := config.ImageSizes
	config.ImageSizes = map[string]config.ImgSize{"xlarge": {ImgWidth: 0}}
	t.Cleanup(func() { config.ImageSizes = previousSizes })

	source := t.TempDir()
	corrupt := filepath.Join(source, "corrupt.jpg")
	var original bytes.Buffer
	original.WriteString("invalid image data")
	if err := os.WriteFile(corrupt, original.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	valid := filepath.Join(source, "valid.jpg")
	var small bytes.Buffer
	if err := jpeg.Encode(&small, image.NewRGBA(image.Rect(0, 0, 20, 20)), nil); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(valid, small.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(source, "missing.jpg")
	photos := []datastore.Picture{
		{Id: "corrupt", Name: "A title that is not the filename", Path: corrupt},
		{Id: "valid", Path: valid},
		{Id: "missing", Path: missing},
	}
	renderer := NewRenderPipeline(&config.GalleryConfiguration{Destpath: t.TempDir()}, nil)
	stat := monitor.NewProgressStats("Building images", len(photos))
	err := NewBatchProcessing(renderer.imageGenV2, photos, stat).Run()
	if err == nil {
		t.Fatal("expected the corrupt and missing photos to fail")
	}
	for _, detail := range []string{"2 items failed:", corrupt, "header decode failed", missing} {
		if !strings.Contains(err.Error(), detail) {
			t.Errorf("batch error is missing %q: %v", detail, err)
		}
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("batch discarded the underlying source error: %v", err)
	}
	if strings.Index(err.Error(), corrupt) > strings.Index(err.Error(), missing) {
		t.Fatalf("failures are not in input order: %v", err)
	}
	snapshot := stat.Snapshot()
	if snapshot.State != monitor.ERROR || snapshot.Processed != len(photos) || snapshot.Message != err.Error() {
		t.Fatalf("activity did not receive the complete failure report: %+v", snapshot)
	}
	if _, err := os.Stat(filepath.Join(renderer.imgDir, "valid", "xlarge.webp")); err != nil {
		t.Fatalf("failed photo stopped the other exports: %v", err)
	}
	files, err := os.ReadDir(filepath.Join(renderer.imgDir, "corrupt"))
	if err != nil || len(files) != 0 {
		t.Fatalf("failed export left a partial file: %v, %v", files, err)
	}
	after, err := os.ReadFile(corrupt)
	if err != nil || !bytes.Equal(after, original.Bytes()) {
		t.Fatalf("source photo was changed: %v", err)
	}
}

func TestProcessImageFitsWebPDimensions(t *testing.T) {
	for _, tc := range []struct {
		name          string
		width, height int
		requested     int
		want          image.Point
	}{
		{"wide", 32766, 2, 0, image.Pt(16383, 1)},
		{"tall", 2, 32766, 0, image.Pt(1, 16383)},
		{"one pixel wide", 1, 16384, 0, image.Pt(1, 16383)},
		{"one pixel tall", 16384, 1, 0, image.Pt(16383, 1)},
		{"at width limit", 16383, 2, 0, image.Pt(16383, 2)},
		{"at height limit", 2, 16383, 0, image.Pt(2, 16383)},
		{"ordinary original", 120, 80, 0, image.Pt(120, 80)},
		{"ordinary resized", 120, 80, 60, image.Pt(60, 40)},
		{"smaller panorama", 32766, 2, 200, image.Pt(200, 1)},
		{"requested width exceeds limit", 200, 1, 20000, image.Pt(16383, 82)},
		{"requested height exceeds limit", 1, 1000, 20, image.Pt(16, 16383)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var encoded bytes.Buffer
			if err := ProcessImage(image.NewNRGBA(image.Rect(0, 0, tc.width, tc.height)), tc.requested, config.WebP, &encoded); err != nil {
				t.Fatal(err)
			}
			decoded, err := webp.Decode(&encoded)
			if err != nil {
				t.Fatalf("invalid WebP export: %v", err)
			}
			if got := decoded.Bounds().Size(); got != tc.want {
				t.Fatalf("export dimensions = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestImageGenerationResizesPanoramaAndPreservesOriginal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "panorama.jpg")
	var original bytes.Buffer
	if err := jpeg.Encode(&original, image.NewRGBA(image.Rect(0, 0, 32766, 4)), nil); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, original.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	photo := datastore.Picture{Id: "panorama", Path: path, Ext: ".jpg"}
	renderer := NewRenderPipeline(&config.GalleryConfiguration{Destpath: t.TempDir(), UseOriginal: true}, nil)
	stat := monitor.NewProgressStats("Building images", 1)
	if err := NewBatchProcessing(renderer.imageGenV2, []datastore.Picture{photo}, stat).Run(); err != nil {
		t.Fatal(err)
	}
	if snapshot := stat.Snapshot(); snapshot.State != monitor.COMPLETE || snapshot.Processed != 1 {
		t.Fatalf("resized image did not complete successfully: %+v", snapshot)
	}
	for name := range config.ImageSizes {
		data, err := os.ReadFile(filepath.Join(renderer.imgDir, photo.Id, name+".webp"))
		if err != nil {
			t.Fatal(err)
		}
		dimensions, err := webp.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		if dimensions.Width > 16383 || dimensions.Height > 16383 {
			t.Fatalf("%s export exceeds WebP limits: %+v", name, dimensions)
		}
		if name == "xlarge" && (dimensions.Width != 16383 || dimensions.Height != 2) {
			t.Fatalf("full-size panorama was not proportionally resized: %+v", dimensions)
		}
	}
	for _, originalPath := range []string{path, filepath.Join(renderer.imgDir, photo.Id, "original.jpg")} {
		data, err := os.ReadFile(originalPath)
		if err != nil || !bytes.Equal(data, original.Bytes()) {
			t.Fatalf("original photo changed at %s: %v", originalPath, err)
		}
	}
}
