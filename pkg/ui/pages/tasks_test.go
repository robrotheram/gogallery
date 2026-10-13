package pages

import (
	"bytes"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"gogallery/pkg/monitor"
	uiMonitor "gogallery/pkg/ui/monitors"
)

func newRescanTestPage(t *testing.T) *TasksPage {
	t.Helper()
	previous := config.Config
	config.Config = &config.Configuration{Gallery: config.GalleryConfiguration{Basepath: t.TempDir()}}
	t.Cleanup(func() { config.Config = previous })
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	m := uiMonitor.NewUIMonitor()
	db, err := datastore.Open(filepath.Join(t.TempDir(), "test.db"), m)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return &TasksPage{DataStore: db, monitor: m}
}

func TestRescanImportsNewPhotosAndGeneratesThumbnails(t *testing.T) {
	page := newRescanTestPage(t)
	if err := page.rescanGallery(); err != nil {
		t.Fatal(err)
	}
	album := filepath.Join(config.Config.Gallery.Basepath, "New album")
	if err := os.Mkdir(album, 0o755); err != nil {
		t.Fatal(err)
	}
	var photo bytes.Buffer
	if err := jpeg.Encode(&photo, image.NewRGBA(image.Rect(0, 0, 8, 8)), nil); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(album, "new.jpg"), photo.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := page.rescanGallery(); err != nil {
		t.Fatal(err)
	}
	pictures, err := page.Pictures.GetAll()
	if err != nil || len(pictures) != 1 {
		t.Fatalf("pictures = %v, error = %v", pictures, err)
	}
	thumbnail, err := page.ImageCache.Get(pictures[0].Id, config.JPEG, "small")
	if err != nil {
		t.Fatal(err)
	}
	defer thumbnail.Close()
	if _, err := jpeg.Decode(thumbnail); err != nil {
		t.Fatalf("invalid thumbnail: %v", err)
	}
	if state := page.monitor.Tasks["Rescan gallery"].Snapshot().State; state != monitor.COMPLETE {
		t.Fatalf("rescan state = %s", state)
	}
}

func TestRescanReportsMissingSource(t *testing.T) {
	page := newRescanTestPage(t)
	config.Config.Gallery.Basepath = ""
	if err := page.rescanGallery(); err == nil {
		t.Fatal("expected missing source error")
	}
	stat := page.monitor.Tasks["Rescan gallery"].Snapshot()
	if stat.State != monitor.ERROR || !strings.Contains(stat.Message, "Settings") {
		t.Fatalf("unexpected rescan result: %s: %s", stat.State, stat.Message)
	}
}
