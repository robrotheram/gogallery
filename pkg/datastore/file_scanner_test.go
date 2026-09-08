package datastore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gogallery/pkg/config"
)

func TestScanRootRejectsBlankPath(t *testing.T) {
	for _, path := range []string{"", "   ", "\t\n"} {
		if _, err := scanRoot(path); err == nil {
			t.Fatalf("scanRoot(%q) unexpectedly succeeded", path)
		}
	}
}

func TestScanRootRejectsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "photo.jpg")
	if err := os.WriteFile(path, []byte("not an image"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := scanRoot(path)
	if err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("expected a not-a-directory error, got %v", err)
	}
}

func TestScanRootReturnsAbsoluteDirectory(t *testing.T) {
	dir := t.TempDir()
	root, err := scanRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(root) {
		t.Fatalf("expected absolute path, got %q", root)
	}
}

func TestWalkPathIgnoresSymbolicLinkImages(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.jpg")
	if err := os.WriteFile(outside, []byte("not really an image"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "linked.jpg")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	pictures, _, _, err := (&DataStore{}).walkPath(root, config.GalleryConfiguration{})
	if err != nil {
		t.Fatal(err)
	}
	if len(pictures) != 0 {
		t.Fatalf("walkPath included symbolic-link images: %#v", pictures)
	}
}
