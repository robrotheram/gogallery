package templateengine

import (
	"bytes"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"gogallery/pkg/datastore"
)

func TestAlbumHeaderBackLink(t *testing.T) {
	engine := NewTemplateEngine()
	if err := engine.LoadFromPath(testThemePath()); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, href, label string
		parent            datastore.Album
	}{
		{name: "top level", href: "/albums", label: "Back to Collections"},
		{name: "nested", href: "/album/parent", label: "Back to Travel", parent: datastore.Album{Id: "parent", Name: "Travel"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			page := Page{Album: datastore.Album{Id: "album", Name: "Italy"}.ToAlbumNode(), ParentAlbum: tc.parent}
			var output bytes.Buffer
			if err := engine.Cache.Get(CollectionTemplate).ExecuteTemplate(&output, "header", page); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(output.String(), `href="`+tc.href+`"`) || !strings.Contains(output.String(), tc.label) {
				t.Fatalf("expected link to %q labelled %q; got %s", tc.href, tc.label, output.String())
			}
		})
	}
}

func testThemePath() string {
	return filepath.Join("..", "..", "themes", "EmeraldNoir")
}

func TestRenderPageReturnsErrorForMissingTemplate(t *testing.T) {
	engine := NewTemplateEngine()
	if err := engine.LoadFromPath(testThemePath()); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	err := engine.RenderPage(&output, "missing", Page{})
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("expected missing-template error, got %v", err)
	}
}

func TestTemplateEngineConcurrentLoadAndRender(t *testing.T) {
	engine := NewTemplateEngine()
	if err := engine.LoadFromPath(testThemePath()); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 5 {
				if err := engine.LoadFromPath(testThemePath()); err != nil {
					t.Errorf("LoadFromPath() error = %v", err)
					return
				}
				var output bytes.Buffer
				if err := engine.RenderPage(&output, HomeTemplate, Page{}); err != nil {
					t.Errorf("RenderPage() error = %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()
}
