package ai

import (
	"errors"
	"fmt"
	"os"
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
	- A clear, descriptive title that names the main subject directly, usually 2-6 words. Prefer simple names such as "Palm House", "Glasshouse Interior", or "Sailing Boats on a Lake". Add a distinguishing detail only when useful. Avoid poetic phrases, vague mood labels, and promotional language such as "A Timeless Escape" or "Nature's Hidden Gem". Use a specific building or place name only when confidently supported by the image; otherwise use a descriptive subject name. Do not pad a short, accurate title to meet a word count. Do not use quotes, emoji, clickbait, or hashtags.
	- A vivid description of 1-2 sentences (roughly 120-180 characters). Lead with the main subject and add useful context about the setting, light, colour, action, or mood. Write for people and search previews, not as a list of keywords.
	- 5-10 concise, distinct tags describing visible subjects, setting, light, and photographic style. Return tags as a JSON array of strings, without hashtags or commas inside individual tags. Avoid speculative locations or identities.
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

// GenerateMetadata returns text suggestions for review. Saving is an explicit
// editor action; the source image and existing photo metadata are never written.
func GenerateMetadata(db *datastore.DataStore, id string) (*ImageMetadata, error) {
	clientsMu.RLock()
	c, ok := clients["gemini"]
	clientsMu.RUnlock()
	if !ok {
		return nil, errors.New("AI suggestions are unavailable. Enable AI and add a Gemini API key in Settings")
	}
	stat := db.NewTask("AI suggestions for "+id, 1)
	defer stat.Complete()
	stat.Start()

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

	return metadata, nil
}
