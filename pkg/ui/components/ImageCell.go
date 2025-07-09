package components

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// ImageCell is a custom widget for a clickable image with a border on hover
type ImageCell struct {
	widget.BaseWidget
	img     *canvas.Image
	border  *canvas.Rectangle
	onClick func()
}

func NewImageCell(img *canvas.Image, onClick func()) *ImageCell {
	border := canvas.NewRectangle(color.NRGBA{R: 0, G: 120, B: 255, A: 100})
	border.StrokeWidth = 4
	border.StrokeColor = color.NRGBA{R: 0, G: 120, B: 255, A: 255}
	border.Hide()
	cell := &ImageCell{
		img:     img,
		border:  border,
		onClick: onClick,
	}
	cell.ExtendBaseWidget(cell)
	return cell
}

func (c *ImageCell) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewStack(c.img, c.border))
}

func (c *ImageCell) Tapped(_ *fyne.PointEvent) {
	if c.onClick != nil {
		c.onClick()
	}
}
func (c *ImageCell) MouseIn(_ *desktop.MouseEvent) {
	c.border.Show()
	c.Refresh()
}

func (c *ImageCell) MouseOut() {
	c.border.Hide()
	c.Refresh()
}
func (c *ImageCell) MouseMoved(_ *desktop.MouseEvent) {
	// This function is intentionally left empty because the ImageCell does not need
	// to handle mouse movement events. It only reacts to mouse hover (MouseIn/MouseOut)
	// and click (Tapped) events.
}

// Set the cursor to a pointer (hand) when hovering over the image cell
func (c *ImageCell) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}
