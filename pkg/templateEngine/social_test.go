package templateengine

import (
	"bytes"
	"strings"
	"testing"

	"gogallery/pkg/config"
)

func TestSocialLinksAndIcons(t *testing.T) {
	engine := NewTemplateEngine()
	if err := engine.LoadFromPath(testThemePath()); err != nil {
		t.Fatal(err)
	}
	tmpl := engine.Cache.Get(HomeTemplate)
	var twitter, facebook bytes.Buffer
	if err := tmpl.ExecuteTemplate(&twitter, "twitter", nil); err != nil {
		t.Fatal(err)
	}
	if err := tmpl.ExecuteTemplate(&facebook, "facebook", nil); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(twitter.String(), "lucide-facebook") || !strings.Contains(twitter.String(), "M18.901 1.153") {
		t.Fatal("X must render its X logo")
	}
	if !strings.Contains(facebook.String(), "lucide-facebook") {
		t.Fatal("Facebook must render its own icon")
	}
	page := Page{Author: config.AboutConfiguration{
		Twitter: "example", Email: "hello@example.com", Facebook: "https://facebook.com/example",
		Website: "https://example.com", Blog: "https://example.com/blog",
	}}
	var output bytes.Buffer
	if err := tmpl.ExecuteTemplate(&output, "social", page); err != nil {
		t.Fatal(err)
	}
	for _, href := range []string{"https://x.com/example", "mailto:hello@example.com", page.Author.Facebook, page.Author.Website} {
		if !strings.Contains(output.String(), `href="`+href+`"`) {
			t.Errorf("missing social link %q", href)
		}
	}
	if !strings.Contains(output.String(), "lucide-earth") {
		t.Fatal("missing website globe icon")
	}
	if strings.Contains(output.String(), `href="`+page.Author.Blog+`"`) {
		t.Fatal("website and blog must use a single combined link")
	}
	page.Author.Blog = page.Author.Website
	output.Reset()
	if err := tmpl.ExecuteTemplate(&output, "social", page); err != nil {
		t.Fatal(err)
	}
	if strings.Count(output.String(), `href="`+page.Author.Website+`"`) != 1 {
		t.Fatal("same website/blog should appear only once")
	}
	page.Author.Website = ""
	output.Reset()
	if err := tmpl.ExecuteTemplate(&output, "social", page); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `href="`+page.Author.Blog+`"`) {
		t.Fatal("legacy blog link is missing")
	}
}
