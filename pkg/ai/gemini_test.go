package ai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"

	"google.golang.org/genai"
)

func TestGenerateMetadataRetriesServiceUnavailable(t *testing.T) {
	for _, tc := range []struct {
		name                       string
		status, failures, attempts int
	}{
		{"recovers on fifth retry", 503, 5, 6},
		{"stops after fifth retry", 503, 6, 6},
		{"does not retry bad request", 400, 1, 1},
		{"does not retry missing model", 404, 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var attempts atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if int(attempts.Add(1)) <= tc.failures {
					w.WriteHeader(tc.status)
					_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"code": tc.status, "message": "Model unavailable"}})
					return
				}
				_, _ = w.Write([]byte(`{"candidates":[{"content":{"role":"model","parts":[{"text":"{\"title\":\"Palm House\",\"description\":\"Palms fill a glasshouse.\",\"tags\":[\"palms\",\"glasshouse\"]}"}]},"finishReason":"STOP"}]}`))
			}))
			defer server.Close()
			client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
				APIKey: "test-key", Backend: genai.BackendGeminiAPI,
				HTTPOptions: genai.HTTPOptions{BaseURL: server.URL, APIVersion: "v1beta"},
			})
			if err != nil {
				t.Fatal(err)
			}
			metadata, err := (&GeminiClient{Client: client}).GenerateMetadata([]byte("\x89PNG\r\n\x1a\n"))
			if got := int(attempts.Load()); got != tc.attempts {
				t.Fatalf("attempts = %d, want %d", got, tc.attempts)
			}
			if tc.failures < tc.attempts {
				if err != nil || metadata == nil || metadata.Title != "Palm House" {
					t.Fatalf("expected successful suggestions, got %+v, %v", metadata, err)
				}
			} else {
				var apiErr genai.APIError
				if !errors.As(err, &apiErr) || apiErr.Code != tc.status {
					t.Fatalf("expected HTTP %d error, got %v", tc.status, err)
				}
			}
		})
	}
}

func TestGenerateMetadataUsesCurrentModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1beta/models/gemini-3.8-flash:generateContent" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var request struct {
			Contents []struct {
				Parts []json.RawMessage `json:"parts"`
			} `json:"contents"`
			GenerationConfig struct {
				ResponseMIMEType string        `json:"responseMimeType"`
				ResponseSchema   *genai.Schema `json:"responseSchema"`
			} `json:"generationConfig"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if len(request.Contents) != 1 || len(request.Contents[0].Parts) != 2 || request.GenerationConfig.ResponseMIMEType != "application/json" {
			t.Errorf("expected image and prompt with structured JSON output, got %+v", request)
		}
		schema := request.GenerationConfig.ResponseSchema
		if schema == nil || schema.Properties["tags"] == nil || schema.Properties["tags"].Type != genai.TypeArray || schema.Properties["tags"].Items.Type != genai.TypeString {
			t.Error("missing tags array schema")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"role":"model","parts":[{"text":"{\"title\":\"Morning Across the Valley\",\"description\":\"Soft light fills a quiet valley.\",\"tags\":[\"valley\",\"morning light\"]}"}]},"finishReason":"STOP"}]}`))
	}))
	defer server.Close()
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey: "test-key", Backend: genai.BackendGeminiAPI,
		HTTPOptions: genai.HTTPOptions{BaseURL: server.URL, APIVersion: "v1beta"},
	})
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := (&GeminiClient{Client: client}).GenerateMetadata([]byte("\x89PNG\r\n\x1a\n"))
	if err != nil {
		t.Fatal(err)
	}
	if metadata.Title != "Morning Across the Valley" || metadata.Description != "Soft light fills a quiet valley." || !reflect.DeepEqual(metadata.Tags, []string{"valley", "morning light"}) {
		t.Fatalf("unexpected suggestions: %+v", metadata)
	}
}

func TestNormalizeMetadata(t *testing.T) {
	metadata := &ImageMetadata{
		Title:       "  Light Across the Valley  ",
		Tags:        []string{" #Landscape ", "landscape", "", "morning   light", "valley, mountains", "Unknown place"},
		Description: "Morning light\n  crosses a quiet mountain valley.",
	}

	if err := normalizeMetadata(metadata); err != nil {
		t.Fatalf("normalizeMetadata() error = %v", err)
	}

	if !reflect.DeepEqual(metadata.Tags, []string{"Landscape", "morning light", "valley", "mountains", "Unknown place"}) {
		t.Fatalf("unexpected tags: %v", metadata.Tags)
	}
	if metadata.Title != "Light Across the Valley" {
		t.Errorf("Title = %q", metadata.Title)
	}
	if metadata.Description != "Morning light crosses a quiet mountain valley." {
		t.Errorf("Description = %q", metadata.Description)
	}
}

func TestNormalizeMetadataRejectsIncompleteResponse(t *testing.T) {
	for _, metadata := range []ImageMetadata{
		{Title: "A title"},
		{Title: "A title", Description: "A description"},
		{Title: "A title", Description: "A description", Tags: []string{" ", "#"}},
	} {
		if err := normalizeMetadata(&metadata); err == nil {
			t.Fatal("normalizeMetadata() returned nil for an incomplete response")
		}
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
