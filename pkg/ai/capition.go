package ai

import (
	"fmt"
	"gogallery/pkg/datastore"
	"os"
)

type ImageCaption struct {
	Title   string `json:"title"`
	Caption string `json:"caption"`
}

const basePrompt = `
	You are a helpful assistant. 
	You will be given an image and I need a title and a caption for it.
	Please provide a short title and a detailed caption for the image. 
	The Captions should be descriptive and engaging, providing context and details about the image.
	Make sure to include any relevant information that would help someone understand the image better.
`

type AIClient interface {
	GenerateCaption(image []byte) (*ImageCaption, error)
}

var clients = map[string]AIClient{}

func IsAi() bool {
	return len(clients) > 0
}

func GenerateCaption(db *datastore.DataStore, id string) (*ImageCaption, error) {
	stat := db.NewTask("Ai Caption for "+id, 1)
	defer stat.Complete()
	stat.Start()
	c, ok := clients["gemini"]
	if !ok {
		stat.Fail("AI client not found")
		return nil, fmt.Errorf("AI client not found")
	}
	pic, err := db.Pictures.FindById(id)
	if err != nil {
		stat.Fail("Failed to find picture: " + err.Error())
		return nil, fmt.Errorf("failed to find picture: %w", err)
	}
	bytes, err := os.ReadFile(pic.Path)
	if err != nil {
		stat.Fail("Failed to read image file: " + err.Error())
		return nil, fmt.Errorf("failed to read image file: %w", err)
	}
	caption, err := c.GenerateCaption(bytes)
	if err != nil {
		stat.Fail("Failed to generate caption: " + err.Error())
		return nil, fmt.Errorf("failed to generate caption: %w", err)
	}
	fmt.Println("Generated Caption:", caption.Caption)
	pic.Caption = caption.Caption
	pic.Name = caption.Title
	if err := db.Pictures.Update(pic.Id, pic); err != nil {
		stat.Fail("Failed to update picture: " + err.Error())
		return nil, fmt.Errorf("failed to update picture: %w", err)
	}
	fmt.Println("Picture updated successfully")
	return caption, nil
}
