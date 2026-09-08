package components

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

func TestResponsiveHeaderUsesFullNavigationWhenItFits(t *testing.T) {
	brand := sizedRectangle(150, 40)
	fullNavigation := sizedRectangle(600, 40)
	compactNavigation := sizedRectangle(140, 40)
	layout := &responsiveHeaderLayout{gap: 16}

	layout.Layout([]fyne.CanvasObject{brand, fullNavigation, compactNavigation}, fyne.NewSize(900, 40))

	if !fullNavigation.Visible() || compactNavigation.Visible() {
		t.Fatal("wide header should show full navigation only")
	}
	if got := fullNavigation.Position().X + fullNavigation.Size().Width; got != 900 {
		t.Fatalf("navigation right edge = %.2f, want 900", got)
	}
}

func TestResponsiveHeaderUsesCompactNavigationBeforeOverlap(t *testing.T) {
	brand := sizedRectangle(150, 40)
	fullNavigation := sizedRectangle(600, 40)
	compactNavigation := sizedRectangle(140, 40)
	layout := &responsiveHeaderLayout{gap: 16}

	layout.Layout([]fyne.CanvasObject{brand, fullNavigation, compactNavigation}, fyne.NewSize(500, 40))

	if fullNavigation.Visible() || !compactNavigation.Visible() {
		t.Fatal("narrow header should show compact navigation only")
	}
	if got := compactNavigation.Position().X + compactNavigation.Size().Width; got != 500 {
		t.Fatalf("compact navigation right edge = %.2f, want 500", got)
	}
}

func sizedRectangle(width, height float32) *canvas.Rectangle {
	rectangle := canvas.NewRectangle(color.Transparent)
	rectangle.SetMinSize(fyne.NewSize(width, height))
	return rectangle
}
