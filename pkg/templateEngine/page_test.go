package templateengine

import (
	"bytes"
	"path/filepath"
	"reflect"
	"testing"

	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
)

func TestSocialSEOFromPhoto(t *testing.T) {
	previous := *config.Config
	t.Cleanup(func() { *config.Config = previous })
	config.Config.Gallery.Name = "My Gallery"
	config.Config.Gallery.Url = "https://photos.example/"
	config.Config.About.Description = "A photo gallery"

	seo := NewSocialSEO("/")
	seo.SetFromPhoto(datastore.Picture{
		Id:        "photo-id",
		Name:      "Light Across the Valley",
		Caption:   "Morning light crosses a quiet mountain valley.",
		Tags:      "landscape, golden hour",
		Dimension: "2048x1365",
	})

	if seo.Site != "https://photos.example/photo/photo-id/" {
		t.Errorf("Site = %q", seo.Site)
	}
	if seo.ImageUrl != "https://photos.example/img/photo-id/xlarge.webp" {
		t.Errorf("ImageUrl = %q", seo.ImageUrl)
	}
	if seo.Type != "article" {
		t.Errorf("Type = %q", seo.Type)
	}
	if !reflect.DeepEqual(seo.Tags, []string{"landscape", "golden hour"}) {
		t.Errorf("Tags = %#v", seo.Tags)
	}
	if seo.ImageWidth != 2048 || seo.ImageHeight != 1365 {
		t.Errorf("image dimensions = %dx%d", seo.ImageWidth, seo.ImageHeight)
	}
}

func TestThemesRenderPhotoMetadata(t *testing.T) {
	previous := *config.Config
	t.Cleanup(func() { *config.Config = previous })
	config.Config.Gallery.Name = "My Gallery"
	config.Config.Gallery.Url = "https://photos.example"
	config.Config.About.Description = "A photo gallery"

	for _, theme := range []string{"EmeraldNoir"} {
		t.Run(theme, func(t *testing.T) {
			engine := NewTemplateEngine()
			if err := engine.LoadFromPath(filepath.Join("..", "..", "themes", theme)); err != nil {
				t.Fatalf("LoadFromPath() error = %v", err)
			}

			page := NewPage(nil)
			picture := datastore.Picture{
				Id:        "photo-id",
				Name:      "Light Across the Valley",
				Caption:   "Morning light crosses a quiet mountain valley.",
				Tags:      "landscape, golden hour",
				Dimension: "2048x1365",
			}
			page.Picture = NewPagePicture(picture)
			page.SEO.SetFromPhoto(picture)

			var output bytes.Buffer
			if err := engine.Cache.Get(PhotoTemplate).Execute(&output, page); err != nil {
				t.Fatalf("render photo template: %v", err)
			}
			if !bytes.Contains(output.Bytes(), []byte(`name="twitter:card" content="summary_large_image"`)) {
				t.Error("rendered page is missing the large Twitter card metadata")
			}
		})
	}
}
