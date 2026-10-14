package preview

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gogallery/pkg/config"
	"gogallery/pkg/datastore"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPreviewPagesServeHTML(t *testing.T) {
	previous := *config.Config
	t.Cleanup(func() { *config.Config = previous })
	config.Config.Gallery.Theme = filepath.Join("..", "..", "themes", "EmeraldNoir")
	config.Config.Gallery.Basepath = t.TempDir()
	if err := os.Mkdir(filepath.Join(config.Config.Gallery.Basepath, "Album"), 0o755); err != nil {
		t.Fatal(err)
	}
	config.Config.Gallery.AlbumBlacklist = nil
	config.Config.Gallery.PictureBlacklist = nil
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&datastore.Picture{}, &datastore.Album{}); err != nil {
		t.Fatal(err)
	}
	store := &datastore.DataStore{Pictures: datastore.NewPictureCollection(db), Albums: datastore.NewAlbumCollection(db)}
	if err := db.Create(&datastore.Picture{Id: "photo", Name: "Photo", Album: "album", AlbumName: "Album", Visibility: "PUBLIC", Path: filepath.Join(config.Config.Gallery.Basepath, "Album", "photo.jpg")}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&datastore.Album{Id: "album", Name: "Album", ProfileId: "photo", ParentPath: config.Config.Gallery.Basepath, Path: filepath.Join(config.Config.Gallery.Basepath, "Album")}).Error; err != nil {
		t.Fatal(err)
	}
	server := NewServer(store)
	for _, path := range []string{"/", "/albums", "/about", "/about/", "/photo/photo", "/album/album"} {
		for _, encoding := range []string{"identity", "gzip"} {
			t.Run(path+"/"+encoding, func(t *testing.T) {
				request := httptest.NewRequest(http.MethodGet, path, nil)
				request.Header.Set("Accept-Encoding", encoding)
				response := httptest.NewRecorder()
				server.ServeHTTP(response, request)
				if response.Code != http.StatusOK {
					t.Fatalf("status = %d, body = %s", response.Code, response.Body)
				}
				if got := response.Result().Header.Get("Content-Type"); got != "text/html; charset=utf-8" {
					t.Errorf("Content-Type = %q, want text/html; charset=utf-8", got)
				}
				var body io.Reader = response.Body
				if encoding == "gzip" {
					reader, err := gzip.NewReader(body)
					if err != nil {
						t.Fatal(err)
					}
					defer reader.Close()
					body = reader
				}
				html, err := io.ReadAll(body)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(string(html), "<!doctype html>") {
					t.Fatalf("response is not an HTML document: %.100s", html)
				}
			})
		}
	}
}

func TestInvalidThemeReturnsServerError(t *testing.T) {
	previous := *config.Config
	t.Cleanup(func() { *config.Config = previous })
	config.Config.Gallery.Theme = t.TempDir()

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	server := &Server{}

	server.PreviewPageHandler(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(response.Body.String(), "Unable to render gallery preview") {
		t.Fatalf("unexpected response body: %q", response.Body.String())
	}
}

func newImageHandlerTestServer(t *testing.T) *Server {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&datastore.Picture{}); err != nil {
		t.Fatal(err)
	}
	return &Server{DataStore: &datastore.DataStore{Pictures: datastore.NewPictureCollection(db)}}
}

func TestImageHandlerReturnsNotFoundForUnknownPicture(t *testing.T) {
	server := newImageHandlerTestServer(t)
	request := httptest.NewRequest(http.MethodGet, "/img/missing/small.webp", nil)
	request = mux.SetURLVars(request, map[string]string{"id": "missing", "size": "small", "ext": "webp"})
	response := httptest.NewRecorder()

	server.ImgHandler(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestImageHandlerRejectsInvalidCacheSize(t *testing.T) {
	server := newImageHandlerTestServer(t)
	if err := server.Pictures.Save(datastore.Picture{
		Id: "photo", Name: "photo", AlbumName: "album", Visibility: "PUBLIC", Path: "/does/not/matter.jpg",
	}); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/img/photo/..%2Fsecret.webp", nil)
	request = mux.SetURLVars(request, map[string]string{"id": "photo", "size": "../secret", "ext": "webp"})
	response := httptest.NewRecorder()

	server.ImgHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestImageHandlerDoesNotExposePrivatePicture(t *testing.T) {
	server := newImageHandlerTestServer(t)
	if err := server.Pictures.Save(datastore.Picture{Id: "private", Visibility: "PRIVATE"}); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/img/private/small.webp", nil)
	request = mux.SetURLVars(request, map[string]string{"id": "private", "size": "small", "ext": "webp"})
	response := httptest.NewRecorder()

	server.ImgHandler(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	handler := securityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := response.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q", got)
	}
}

func TestPreviewServerUsesRequestTimeouts(t *testing.T) {
	previous := config.Config.UI.Public
	config.Config.UI.Public = false
	t.Cleanup(func() { config.Config.UI.Public = previous })
	server := NewServer(nil)
	if err := server.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Stop() })

	if server.server.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("ReadHeaderTimeout = %v", server.server.ReadHeaderTimeout)
	}
	if server.server.IdleTimeout != 60*time.Second {
		t.Fatalf("IdleTimeout = %v", server.server.IdleTimeout)
	}
}

func TestPreviewProfilePhotoServesOnlyConfiguredFile(t *testing.T) {
	previous := config.Config.About
	t.Cleanup(func() { config.Config.About = previous })
	path := filepath.Join(t.TempDir(), "portrait.png")
	data := []byte("\x89PNG\r\n\x1a\nprofile image")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	config.Config.About.ProfilePhoto = path
	server := NewServer(nil)
	for _, tc := range []struct {
		path   string
		status int
	}{{"/about/profile.png", 200}, {"/about/other.png", 404}} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, tc.path, nil)
		request.Header.Set("Accept-Encoding", "gzip")
		server.ServeHTTP(response, request)
		if response.Code != tc.status {
			t.Fatalf("%s: status %d", tc.path, response.Code)
		}
		if tc.status == 200 && (response.Body.String() != string(data) || response.Header().Get("Content-Type") != "image/png") {
			t.Fatal("profile photo response is incorrect")
		}
	}
}
