package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"image/color"
	"testing"
)

func sizedRectangle(width, height float32) *canvas.Rectangle {
	r := canvas.NewRectangle(color.Transparent)
	r.SetMinSize(fyne.NewSize(width, height))
	return r
}
func TestShellNavigationReflowsWithoutOverlap(t *testing.T) {
	side, compact, content := sizedRectangle(188, 300), sizedRectangle(200, 50), sizedRectangle(360, 240)
	objects := []fyne.CanvasObject{side, compact, content}
	l := &shellLayout{}
	for _, width := range []float32{1400, 900, 899, 520, 1200} {
		l.Layout(objects, fyne.NewSize(width, 600))
		if width >= 900 {
			if !side.Visible() || compact.Visible() {
				t.Fatal("desktop should use sidebar navigation")
			}
			if content.Position().X != 188 || content.Size().Width != width-188 {
				t.Fatalf("content overlaps sidebar at %v", width)
			}
		} else {
			if side.Visible() || !compact.Visible() {
				t.Fatal("small windows should use toolbar navigation")
			}
			if content.Position().Y != 50 || content.Size().Width != width {
				t.Fatalf("content overlaps toolbar at %v", width)
			}
		}
	}
}
