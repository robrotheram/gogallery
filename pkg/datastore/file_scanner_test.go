package datastore

import (
	"gogallery/pkg/monitor"
	"image"
	"image/png"
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

func TestScanPreservesAlbumEditsAndRepairsDeletedCover(t *testing.T) {
	root := t.TempDir()
	previous := config.Config
	config.Config = &config.Configuration{Gallery: config.GalleryConfiguration{Basepath: root}}
	t.Cleanup(func() { config.Config = previous })
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	db, err := Open(filepath.Join(t.TempDir(), "gallery.db"), monitor.NewMonitor())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	folder := filepath.Join(root, "Original folder")
	if err = os.Mkdir(folder, 0700); err != nil {
		t.Fatal(err)
	}
	writePhoto := func(name string) string {
		t.Helper()
		path := filepath.Join(folder, name)
		file, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if err = png.Encode(file, image.NewRGBA(image.Rect(0, 0, 4, 4))); err != nil {
			t.Fatal(err)
		}
		if err = file.Close(); err != nil {
			t.Fatal(err)
		}
		return path
	}
	coverPath := writePhoto("a.png")
	writePhoto("b.png")
	if err = db.ScanPath(root); err != nil {
		t.Fatal(err)
	}
	album, err := db.Albums.FindById(config.GetMD5Hash(folder))
	if err != nil {
		t.Fatal(err)
	}
	album.Name = "A new title"
	album.ProfileId = config.GetMD5Hash(coverPath)
	if err = db.Albums.Update(album.Id, album); err != nil {
		t.Fatal(err)
	}
	// A new import must inherit the edited title as well.
	writePhoto("c.png")
	if err = db.ScanPath(root); err != nil {
		t.Fatal(err)
	}
	got, err := db.Albums.FindById(album.Id)
	if err != nil || got.Name != album.Name || got.ProfileId != album.ProfileId {
		t.Fatalf("scan lost edits: %#v, %v", got, err)
	}
	pictures, err := db.Pictures.GetAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(pictures) != 3 {
		t.Fatalf("imported %d photos, want 3", len(pictures))
	}
	for _, pic := range pictures {
		if pic.AlbumName != album.Name {
			t.Fatalf("photo kept old album title: %#v", pic)
		}
	}
	if err = os.Remove(coverPath); err != nil {
		t.Fatal(err)
	}
	if err = db.ScanPath(root); err != nil {
		t.Fatal(err)
	}
	got, err = db.Albums.FindById(album.Id)
	if err != nil || got.ProfileId == album.ProfileId || got.ProfileId == "" {
		t.Fatalf("deleted cover was not replaced: %#v, %v", got, err)
	}
	if _, err = db.Pictures.FindByID(got.ProfileId); err != nil {
		t.Fatalf("replacement cover does not exist: %v", err)
	}
	if got.Name != album.Name {
		t.Fatal("cover repair lost the edited title")
	}
}
