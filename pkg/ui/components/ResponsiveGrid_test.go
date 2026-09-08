package components

import (
	"image/color"
	"math"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

func TestResponsiveGridFillsAvailableWidthAcrossWindowSizes(t *testing.T) {
	layout := NewResponsiveGridLayoutWithFooter(230, 1.5, 8, 48)

	for _, test := range []struct {
		name    string
		width   float32
		columns int
	}{
		{name: "wide", width: 1261, columns: 5},
		{name: "desktop", width: 900, columns: 3},
		{name: "narrow", width: 640, columns: 2},
		{name: "single column", width: 400, columns: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			columns, cell := layout.gridMetrics(test.width)
			if columns != test.columns {
				t.Fatalf("columns = %d, want %d", columns, test.columns)
			}

			rightEdge := float32(8) + float32(columns-1)*(cell.Width+8) + cell.Width
			if difference := float64(rightEdge - (test.width - 8)); math.Abs(difference) > 0.01 {
				t.Fatalf("right edge = %.2f, want %.2f", rightEdge, test.width-8)
			}
		})
	}
}

func TestResponsiveGridReflowsRowsAfterResize(t *testing.T) {
	layout := NewResponsiveGridLayoutWithFooter(230, 1.5, 8, 48)
	objects := make([]fyne.CanvasObject, 10)
	for i := range objects {
		objects[i] = canvas.NewRectangle(color.Transparent)
	}

	layout.Layout(objects, fyne.NewSize(900, 800))
	wideHeight := layout.MinSize(objects).Height
	layout.Layout(objects, fyne.NewSize(640, 800))
	narrowHeight := layout.MinSize(objects).Height

	if narrowHeight <= wideHeight {
		t.Fatalf("narrow grid height %.2f should exceed wide grid height %.2f", narrowHeight, wideHeight)
	}
	if objects[2].Position().Y <= objects[1].Position().Y {
		t.Fatal("third item did not wrap onto the next row at narrow width")
	}
}
