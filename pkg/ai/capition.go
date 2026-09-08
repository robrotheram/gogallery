package ai

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"gogallery/pkg/datastore"
)

type ImageMetadata struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

const maxAIImageBytes int64 = 20 << 20

const basePrompt = `
	You are an expert photography editor creating accurate, accessible metadata for a public photo gallery.
	Study the image and return:
	- A distinctive title of 3-8 words. Make it natural and memorable, without quotes, emoji, clickbait, or hashtags.
	- A vivid description of 1-2 sentences (roughly 120-180 characters). Lead with the main subject and add useful context about the setting, light, colour, action, or mood. Write for people and search previews, not as a list of keywords.
	- 6-10 concise, distinct tags that improve discovery. Mix concrete subjects with relevant setting, genre, colour, season, action, or mood. Use lowercase words or short phrases without # symbols.
	Only state what the image supports. Do not invent an exact place, event, relationship, identity, or backstory. Do not identify a person unless their identity is unmistakable from widely known public context.
`

type AIClient interface {
	GenerateMetadata(image []byte) (*ImageMetadata, error)
}

var (
	clientsMu sync.RWMutex
	clients   = map[string]AIClient{}
)

func IsAIEnabled() bool {
	clientsMu.RLock()
	defer clientsMu.RUnlock()
	return len(clients) > 0
}

func GenerateMetadata(db *datastore.DataStore, id string) (*ImageMetadata, error) {
	stat := db.NewTask("AI metadata for "+id, 1)
	defer stat.Complete()
	stat.Start()

	clientsMu.RLock()
	c, ok := clients["gemini"]
	clientsMu.RUnlock()
	if !ok {
		stat.Fail("AI client not found")
		return nil, errors.New("AI client not found")
	}

	pic, err := db.Pictures.FindByID(id)
	if err != nil {
		stat.Fail("Failed to find picture: " + err.Error())
		return nil, fmt.Errorf("failed to find picture: %w", err)
	}
	info, err := os.Stat(pic.Path)
	if err != nil {
		stat.Fail("Failed to inspect image file: " + err.Error())
		return nil, fmt.Errorf("inspect image file: %w", err)
	}
	if info.Size() > maxAIImageBytes {
		err := fmt.Errorf("image is too large for AI metadata (%d MB maximum)", maxAIImageBytes>>20)
		stat.Fail(err.Error())
		return nil, err
	}
	imageBytes, err := os.ReadFile(pic.Path)
	if err != nil {
		stat.Fail("Failed to read image file: " + err.Error())
		return nil, fmt.Errorf("failed to read image file: %w", err)
	}
	metadata, err := c.GenerateMetadata(imageBytes)
	if err != nil {
		stat.Fail("Failed to generate metadata: " + err.Error())
		return nil, fmt.Errorf("failed to generate metadata: %w", err)
	}

	pic.Name = metadata.Title
	pic.Caption = metadata.Description
	pic.Tags = strings.Join(metadata.Tags, ", ")
	if err := db.Pictures.Update(pic.Id, pic); err != nil {
		stat.Fail("Failed to update picture: " + err.Error())
		return nil, fmt.Errorf("failed to update picture: %w", err)
	}
	return metadata, nil
}
