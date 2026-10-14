package components

import (
	"fmt"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
)

func TestImageGridPageSizeFollowsViewport(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	previous := *config.Config
	t.Cleanup(func() { *config.Config = previous })
	config.Config.UI.ImagesPerPage = 1
	config.Config.Gallery.ImagesPerPage = 7
	grid := NewImageGrid(nil)
	pictures := make([]datastore.Picture, 200)
	for i := range pictures {
		pictures[i].Name = fmt.Sprintf("Photo %03d", i)
	}
	grid.SetImages(pictures)
	window := test.NewWindow(grid.Layout())
	defer window.Close()
	window.Resize(fyne.NewSize(800, 600))
	small := grid.itemsPerPage
	if small <= 1 {
		t.Fatalf("desktop still uses the saved limit: %d", small)
	}
	window.Resize(fyne.NewSize(1400, 1000))
	large := grid.itemsPerPage
	if large <= small {
		t.Fatalf("larger viewport has %d photos, smaller had %d", large, small)
	}
	if len(grid.currentImages()) != large || len(grid.grid.content.Objects) != large {
		t.Fatal("visible tiles did not follow viewport capacity")
	}
	grid.next.OnTapped()
	first := grid.currentPage * grid.itemsPerPage
	window.Resize(fyne.NewSize(800, 600))
	start := grid.currentPage * grid.itemsPerPage
	if first < start || first >= start+grid.itemsPerPage {
		t.Fatal("resize lost the current position")
	}
	if config.Config.Gallery.ImagesPerPage != 7 {
		t.Fatal("resize changed website pagination")
	}
	grid.search.SetText("Photo 199")
	if grid.currentPage != 0 || len(grid.currentImages()) != 1 {
		t.Fatal("filter did not reset pagination")
	}
}
