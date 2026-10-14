package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"gogallery/pkg/config"

	"google.golang.org/genai"
)

type GeminiClient struct {
	*genai.Client
}

const geminiMetadataModel = "gemini-3.8-flash"

func RegisterGeminiClient() (*GeminiClient, error) {
	ClearGeminiClient()
	if !config.Config.UI.AIEnabled {
		return nil, errors.New("AI suggestions are disabled")
	}
	apiKey := strings.TrimSpace(config.Config.UI.GeminiApiKey)
	if apiKey == "" {
		return nil, errors.New("AI API key is not configured")
	}

	ctx := context.Background()
	cc := &genai.ClientConfig{
		APIKey: apiKey,
	}
	client, err := genai.NewClient(ctx, cc)
	if err != nil {
		log.Printf("Error creating Gemini client: %v", err)
		return nil, err
	}
	gm := &GeminiClient{Client: client}
	clientsMu.Lock()
	clients["gemini"] = gm
	clientsMu.Unlock()
	return gm, nil
}

func ClearGeminiClient() {
	clientsMu.Lock()
	delete(clients, "gemini")
	clientsMu.Unlock()
}

func (g *GeminiClient) GenerateMetadata(image []byte) (*ImageMetadata, error) {
	if len(image) == 0 {
		return nil, errors.New("image is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	generateConfig := &genai.GenerateContentConfig{
		HTTPOptions: &genai.HTTPOptions{
			RetryOptions: &genai.HTTPRetryOptions{
				// Six attempts means the initial request plus five retries.
				Attempts:        genai.Ptr[int32](6),
				InitialDelay:    genai.Ptr(1.0),
				MaxDelay:        genai.Ptr(8.0),
				ExpBase:         genai.Ptr(2.0),
				Jitter:          genai.Ptr(0.0),
				HTTPStatusCodes: []int32{http.StatusServiceUnavailable},
			},
		},
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"title": {
					Type:        genai.TypeString,
					Description: "A clear, concise subject name, usually 2-6 words, such as Palm House or Glasshouse Interior. Prefer concrete descriptions over poetic titles; use specific place names only when supported by the image.",
				},
				"description": {
					Type:        genai.TypeString,
					Description: "An accurate, engaging 1-2 sentence description suitable for search and social previews.",
				},
				"tags": {
					Type:        genai.TypeArray,
					Items:       &genai.Schema{Type: genai.TypeString},
					Description: "5-10 distinct, concise tags grounded in visible image content, without hashtags.",
				},
			},
			PropertyOrdering: []string{"title", "description", "tags"},
			Required:         []string{"title", "description", "tags"},
		},
	}
	mimeType := http.DetectContentType(image)
	if !supportedImageMIME(mimeType) {
		return nil, fmt.Errorf("unsupported image type %q", mimeType)
	}
	parts := []*genai.Part{
		genai.NewPartFromText(basePrompt),
		genai.NewPartFromBytes(image, mimeType),
	}

	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	result, err := g.Client.Models.GenerateContent(
		ctx,
		geminiMetadataModel,
		contents,
		generateConfig,
	)
	if err != nil {
		return nil, err
	}

	var metadata ImageMetadata
	if err := json.Unmarshal([]byte(result.Text()), &metadata); err != nil {
		return nil, fmt.Errorf("decode Gemini response: %w", err)
	}
	if err := normalizeMetadata(&metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}

func normalizeMetadata(metadata *ImageMetadata) error {
	metadata.Title = strings.TrimSpace(metadata.Title)
	metadata.Description = strings.Join(strings.Fields(metadata.Description), " ")
	var tags []string
	seen := make(map[string]bool)
	for _, value := range metadata.Tags {
		for _, tag := range strings.Split(value, ",") {
			tag = strings.Join(strings.Fields(strings.TrimLeft(strings.TrimSpace(tag), "#")), " ")
			key := strings.ToLower(tag)
			if tag != "" && !seen[key] && len(tags) < 10 {
				seen[key] = true
				tags = append(tags, tag)
			}
		}
	}
	metadata.Tags = tags
	if metadata.Title == "" || metadata.Description == "" || len(metadata.Tags) == 0 {
		return errors.New("AI returned incomplete metadata")
	}
	return nil
}

func supportedImageMIME(mimeType string) bool {
	switch mimeType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return true
	default:
		return false
	}
}
