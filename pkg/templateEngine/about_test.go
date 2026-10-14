package templateengine

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"gogallery/pkg/config"
	"golang.org/x/net/html"
)

func TestAboutDescriptionPreservesSavedSpacing(t *testing.T) {
	description := "\nFirst paragraph.\n\nSecond  paragraph.\n  Indented line.\n\tTabbed line.\n<script>text, not markup</script>\n"
	for _, theme := range []string{"EmeraldNoir", "Eastnor"} {
		t.Run(theme, func(t *testing.T) {
			engine := NewTemplateEngine()
			if err := engine.LoadFromPath(filepath.Join("..", "..", "themes", theme)); err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			if err := engine.RenderPage(&output, AboutTemplate, Page{Author: config.AboutConfiguration{Description: description}}); err != nil {
				t.Fatal(err)
			}
			document, err := html.Parse(&output)
			if err != nil {
				t.Fatal(err)
			}
			var got strings.Builder
			var collect func(*html.Node, bool)
			collect = func(node *html.Node, inDescription bool) {
				inDescription = inDescription || (node.Type == html.ElementNode && node.Data == "pre")
				if inDescription && node.Type == html.ElementNode && node.Data == "script" {
					t.Error("description was interpreted as HTML")
				}
				if inDescription && node.Type == html.TextNode {
					got.WriteString(node.Data)
				}
				for child := node.FirstChild; child != nil; child = child.NextSibling {
					collect(child, inDescription)
				}
			}
			collect(document, false)
			if got.String() != description {
				t.Fatalf("description = %q, want %q", got.String(), description)
			}
		})
	}
}
