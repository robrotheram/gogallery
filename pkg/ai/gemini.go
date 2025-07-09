package ai

import (
	"context"
	"encoding/json"
	"gogallery/pkg/config"
	"log"

	"google.golang.org/genai"
)

type GeminiClient struct {
	*genai.Client
}

func RegisterGeminiClient() (*GeminiClient, error) {
	ctx := context.Background()
	cc := &genai.ClientConfig{
		APIKey: config.Config.UI.GeminiApiKey,
	}
	client, err := genai.NewClient(ctx, cc)
	if err != nil {
		log.Printf("Error creating Gemini client: %v", err)
		return nil, err
	}
	gm := &GeminiClient{Client: client}
	clients["gemini"] = gm
	return gm, nil
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
