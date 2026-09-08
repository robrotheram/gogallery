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

func RegisterGeminiClient() (*GeminiClient, error) {
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

	minTags, maxTags := int64(6), int64(10)
	generateConfig := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"title": {
					Type:        genai.TypeString,
					Description: "A memorable, accurate photo title of 3-8 words.",
				},
				"description": {
					Type:        genai.TypeString,
					Description: "An accurate, engaging 1-2 sentence description suitable for search and social previews.",
				},
				"tags": {
					Type:        genai.TypeArray,
					Description: "Distinct lowercase discovery tags without hash symbols.",
					Items:       &genai.Schema{Type: genai.TypeString},
					MinItems:    &minTags,
					MaxItems:    &maxTags,
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
		"gemini-2.5-flash",
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
	metadata.Tags = normalizeTags(metadata.Tags)
	if metadata.Title == "" || metadata.Description == "" {
		return errors.New("AI returned incomplete metadata")
	}
	if len(metadata.Tags) == 0 {
		return errors.New("AI returned no usable tags")
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

func normalizeTags(tags []string) []string {
	normalized := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		tag = strings.ToLower(strings.TrimSpace(strings.TrimLeft(tag, "#")))
		tag = strings.Join(strings.Fields(tag), " ")
		if tag == "" {
			continue
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		normalized = append(normalized, tag)
		if len(normalized) == 10 {
			break
		}
	}
	return normalized
}
