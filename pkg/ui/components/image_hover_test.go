package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"gogallery/pkg/datastore"
	"image"
	"testing"
)

func TestPhotoTitleAppearsOnlyOnHoverWithoutResizingImage(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	tile := NewImage(nil, datastore.Picture{}, nil, nil)
	tile.pic = datastore.Picture{Id: "photo", Name: "A photo title"}
	tile.imageObj = canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 30, 20)))
	tile.content = container.NewStack(tile.imageObj)
	tile.updateCaption(tile.pic)
	tile.SetHoverCaption(true)
	tile.Resize(fyne.NewSize(300, 200))
	before := tile.imageObj.Size()
	if tile.captionOverlay.Visible() {
		t.Fatal("title should start hidden")
	}
	tile.MouseIn(&desktop.MouseEvent{})
	if !tile.captionOverlay.Visible() {
		t.Fatal("hover did not reveal title")
	}
	if tile.imageObj.Size() != before || tile.MinSize() != imageTileMinSize {
		t.Fatal("hover changed the photo or grid dimensions")
	}
	if tile.captionOverlay.Position().Y+tile.captionOverlay.Size().Height != 200 {
		t.Fatal("title should overlay the bottom of the photo")
	}
	tile.SetPicture(datastore.Picture{Id: "photo", Name: "Edited title"})
	if got := tile.captionLabel.Segments[0].(*widget.TextSegment).Text; got != "Edited title" {
		t.Fatalf("hover shows stale title: %q", got)
	}
	tile.MouseOut()
	if tile.captionOverlay.Visible() {
		t.Fatal("title remained after leaving the photo")
	}
	tile.SetActive(true)
	if tile.captionOverlay.Visible() {
		t.Fatal("selection should not permanently show the title")
	}
}
func TestReusedPhotoTileUpdatesHoverTitleWhileLoading(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	tile := NewImage(nil, datastore.Picture{}, nil, nil)
	tile.SetHoverCaption(true)
	tile.MouseIn(&desktop.MouseEvent{})
	// Simulate an outstanding load while the grid reuses the tile.
	tile.isLoading = true
	tile.SetPicture(datastore.Picture{Id: "next", Path: "/photos/next.jpg"})
	if !tile.captionOverlay.Visible() {
		t.Fatal("loading photo should still have a hover title")
	}
	if got := tile.captionLabel.Segments[0].(*widget.TextSegment).Text; got != "next.jpg" {
		t.Fatalf("filename fallback = %q", got)
	}
	tile.SetHoverCaption(false)
	if tile.captionOverlay.Visible() {
		t.Fatal("album covers must not expose photo titles")
	}
}
