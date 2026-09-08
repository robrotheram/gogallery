package preview

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gogallery/pkg/config"
	"gogallery/pkg/datastore"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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
