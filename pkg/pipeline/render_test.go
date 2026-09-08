package pipeline

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"gogallery/pkg/config"
)

func TestRenderPipelinesKeepIndependentOutputPaths(t *testing.T) {
	first := NewRenderPipeline(&config.GalleryConfiguration{Destpath: "/tmp/first"}, nil)
	second := NewRenderPipeline(&config.GalleryConfiguration{Destpath: "/tmp/second"}, nil)

	if first.root == second.root || first.imgDir == second.imgDir {
		t.Fatal("render pipelines unexpectedly share output paths")
	}
}

func TestGeneratedFileIsNotReplacedWhenRenderingFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.html")
	if err := os.WriteFile(path, []byte("previous version"), 0o644); err != nil {
		t.Fatal(err)
	}
	wantErr := errors.New("render failed")
	if err := writeGeneratedFile(path, func(w io.Writer) error {
		_, _ = w.Write([]byte("partial version"))
		return wantErr
	}); !errors.Is(err, wantErr) {
		t.Fatalf("writeGeneratedFile error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "previous version" {
		t.Fatalf("existing output was replaced with %q", data)
	}
}

func TestDeleteSiteRejectsUnsafeDestination(t *testing.T) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for _, destination := range []string{"", workingDirectory, string(os.PathSeparator)} {
		renderer := NewRenderPipeline(&config.GalleryConfiguration{Destpath: destination}, nil)
		if err := renderer.DeleteSite(); err == nil {
			t.Fatalf("DeleteSite accepted unsafe destination %q", destination)
		}
	}
}

func TestDeleteSiteRejectsDestinationOverlappingGallerySource(t *testing.T) {
	galleryRoot := t.TempDir()
	for _, destination := range []string{
		filepath.Join(galleryRoot, "generated-site"),
		filepath.Dir(galleryRoot),
	} {
		renderer := NewRenderPipeline(&config.GalleryConfiguration{
			Basepath: galleryRoot,
			Destpath: destination,
		}, nil)
		if err := renderer.DeleteSite(); err == nil {
			t.Fatalf("DeleteSite accepted destination %q overlapping gallery %q", destination, galleryRoot)
		}
	}
}

func TestDeleteSiteRemovesOnlyConfiguredOutput(t *testing.T) {
	parent := t.TempDir()
	destination := filepath.Join(parent, "site")
	if err := os.MkdirAll(destination, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destination, "index.html"), []byte("site"), 0o644); err != nil {
		t.Fatal(err)
	}

	renderer := NewRenderPipeline(&config.GalleryConfiguration{Destpath: destination}, nil)
	if err := renderer.DeleteSite(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatalf("destination still exists after deletion: %v", err)
	}
	if _, err := os.Stat(parent); err != nil {
		t.Fatalf("parent directory was affected: %v", err)
	}
}
