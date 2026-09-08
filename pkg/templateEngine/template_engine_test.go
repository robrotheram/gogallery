package templateengine

import (
	"bytes"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

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
