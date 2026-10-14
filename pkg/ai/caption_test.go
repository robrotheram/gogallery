package ai

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"gogallery/pkg/monitor"
)

type suggestionClient struct{ received []byte }

func (c *suggestionClient) GenerateMetadata(image []byte) (*ImageMetadata, error) {
	c.received = append([]byte(nil), image...)
	return &ImageMetadata{Title: "Morning in the garden", Description: "Sunlight falls across the garden.", Tags: []string{"garden", "sunlight"}}, nil
}

func TestSuggestionsDoNotModifyThePhotoOrStoredMetadata(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	root := t.TempDir()
	db, err := datastore.Open(filepath.Join(root, "photos.db"), monitor.NewMonitor())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	path := filepath.Join(root, "photo.jpg")
	original := []byte("original photo bytes for a fake AI client")
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}
	picture := datastore.Picture{Id: "photo", Name: "My title", Caption: "My description", Tags: "my, tags", Path: path, Visibility: "PRIVATE"}
	if err := db.Pictures.Save(picture); err != nil {
		t.Fatal(err)
	}
	before, err := db.Pictures.FindByID(picture.Id)
	if err != nil {
		t.Fatal(err)
	}
	client := &suggestionClient{}
	clientsMu.Lock()
	clients["gemini"] = client
	clientsMu.Unlock()
	t.Cleanup(ClearGeminiClient)
	metadata, err := GenerateMetadata(db, picture.Id)
	if err != nil {
		t.Fatal(err)
	}
	if metadata.Title != "Morning in the garden" || !bytes.Equal(client.received, original) || !reflect.DeepEqual(metadata.Tags, []string{"garden", "sunlight"}) {
		t.Fatal("suggestion request did not return the expected text")
	}
	after, err := db.Pictures.FindByID(picture.Id)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("suggestions modified stored metadata before the user saved")
	}
	contents, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(contents, original) {
		t.Fatal("suggestions changed the source image")
	}
	ClearGeminiClient()
	if _, err := GenerateMetadata(nil, picture.Id); err == nil {
		t.Fatal("disabled AI must reject requests before accessing photos")
	}
}

func TestGeminiRegistrationRequiresEnabledPreference(t *testing.T) {
	previous := config.Config
	config.Config = &config.Configuration{UI: config.UIConfiguration{GeminiApiKey: "saved-test-key"}}
	t.Cleanup(func() { config.Config = previous; ClearGeminiClient() })
	if _, err := RegisterGeminiClient(); err == nil || IsAIEnabled() {
		t.Fatal("a stored API key enabled AI without opt-in")
	}
	config.Config.UI.AIEnabled = true
	if _, err := RegisterGeminiClient(); err != nil || !IsAIEnabled() {
		t.Fatalf("explicit opt-in failed to register client: %v", err)
	}
	config.Config.UI.AIEnabled = false
	if _, err := RegisterGeminiClient(); err == nil || IsAIEnabled() {
		t.Fatal("disabled preference left a previously registered client active")
	}
}
