package components

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// ClickableTitle is a canvas.Text that acts like a clickable title
type ClickableTitle struct {
	widget.BaseWidget
	Text     string
	OnTapped func()
}

func NewClickableTitle(text string, onTapped func()) *ClickableTitle {
	c := &ClickableTitle{Text: text, OnTapped: onTapped}
	c.ExtendBaseWidget(c)
	return c
}

func (c *ClickableTitle) CreateRenderer() fyne.WidgetRenderer {
	title := canvas.NewText(c.Text, nil)
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 28
	objs := []fyne.CanvasObject{title}
	return &clickableTitleRenderer{title: title, objects: objs}
}

func (c *ClickableTitle) Tapped(_ *fyne.PointEvent) {
	if c.OnTapped != nil {
		c.OnTapped()
	}
}

type clickableTitleRenderer struct {
	title   *canvas.Text
	objects []fyne.CanvasObject
}

func (r *clickableTitleRenderer) Layout(size fyne.Size) {
	r.title.Resize(size)
}
func (r *clickableTitleRenderer) MinSize() fyne.Size {
	return r.title.MinSize()
}
func (r *clickableTitleRenderer) Refresh() {
	canvas.Refresh(r.title)
}
func (r *clickableTitleRenderer) BackgroundColor() color.Color {
	return color.Transparent
}
func (r *clickableTitleRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}
func (r *clickableTitleRenderer) Destroy() {
	// The Destroy method is intentionally left empty because
	// there are no resources to clean up for this renderer.
}

func (r *ClickableTitle) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}
