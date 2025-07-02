package ai

import (
	"context"
	"encoding/json"
	"log"

	"google.golang.org/genai"
)

type GeminiClient struct {
	*genai.Client
}

func init() {
	// Initialize the Gemini client
	c, err := NewGeminiClient()
	if err != nil {
		log.Fatal("Failed to create Gemini client:", err)
	}
	clients["gemini"] = c
	log.Println("Gemini client initialized successfully")
}

func NewGeminiClient() (*GeminiClient, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	return &GeminiClient{Client: client}, nil
}

func (g *GeminiClient) GenerateCaption(bytes []byte) (*ImageCaption, error) {
	var caption ImageCaption
	ctx := context.Background()

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"title":   {Type: genai.TypeString},
				"caption": {Type: genai.TypeString},
			},
		},
	}
	parts := []*genai.Part{
		genai.NewPartFromText(basePrompt),
		genai.NewPartFromBytes(bytes, "image/jpeg"),
	}

	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	result, err := g.Client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		contents,
		config,
	)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(result.Text()), &caption)
	if err != nil {
		return nil, err
	}
	return &caption, nil
}
