package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gogallery/pkg/config"
	templateengine "gogallery/pkg/templateEngine"
)

func TestRenderAboutIncludesProfileAndCopiesPhoto(t *testing.T) {
	previous := *config.Config
	t.Cleanup(func() { *config.Config = previous })
	photo := filepath.Join(t.TempDir(), "portrait.png")
	if err := os.WriteFile(photo, []byte("profile image"), 0o600); err != nil {
		t.Fatal(err)
	}
	config.Config.About = config.AboutConfiguration{Photographer: "Alex", Description: "Travel & nature photography.\nBased in London.", ProfilePhoto: photo, Website: "https://example.com"}
	for _, theme := range []string{"EmeraldNoir", "Eastnor"} {
		t.Run(theme, func(t *testing.T) {
			if err := templateengine.Templates.LoadFromPath(filepath.Join("..", "..", "themes", theme)); err != nil {
				t.Fatal(err)
			}
			cfg := config.GalleryConfiguration{Destpath: t.TempDir()}
			if err := NewRenderPipeline(&cfg, nil).renderAbout(); err != nil {
				t.Fatal(err)
			}
			page, err := os.ReadFile(filepath.Join(cfg.Destpath, "about", "index.html"))
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"About Alex", "nature photography.", "/about/profile.png"} {
				if !strings.Contains(string(page), want) {
					t.Errorf("About page missing %q", want)
				}
			}
			if strings.Contains(string(page), "https://example.com") {
				t.Fatal("About page must not include the website link")
			}
			copied, err := os.ReadFile(filepath.Join(cfg.Destpath, "about", "profile.png"))
			if err != nil || string(copied) != "profile image" {
				t.Fatalf("profile copy = %q, %v", copied, err)
			}
			if strings.Contains(string(page), photo) {
				t.Fatal("local file path leaked into generated page")
			}
		})
	}
}
