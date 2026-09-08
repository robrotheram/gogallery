package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ResponsiveGridLayout fills the available width with equal-width cells. It
// adds or removes columns at the configured minimum width and stretches each
// row to consume any remainder instead of leaving a blank strip on the right.
type ResponsiveGridLayout struct {
	minCellWidth int
	aspectRatio  float64
	gap          int
	footerHeight float32
	lastWidth    float32
}

func NewResponsiveGridLayout(minCellWidth int, aspectRatio float64, gap int) *ResponsiveGridLayout {
	return NewResponsiveGridLayoutWithFooter(minCellWidth, aspectRatio, gap, 0)
}

func NewResponsiveGridLayoutWithFooter(minCellWidth int, aspectRatio float64, gap int, footerHeight float32) *ResponsiveGridLayout {
	return &ResponsiveGridLayout{
		minCellWidth: minCellWidth,
		aspectRatio:  aspectRatio,
		gap:          gap,
		footerHeight: footerHeight,
	}
}

func (l *ResponsiveGridLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	l.lastWidth = size.Width
	cols, cellSize := l.gridMetrics(size.Width)
	gap := float32(l.gap)

	for i, object := range objects {
		row := i / cols
		col := i % cols
		object.Resize(cellSize)
		object.Move(fyne.NewPos(
			gap+float32(col)*(cellSize.Width+gap),
			gap+float32(row)*(cellSize.Height+gap),
		))
	}
}

func (l *ResponsiveGridLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	width := l.lastWidth
	if width <= 0 {
		width = float32(l.minCellWidth + 2*l.gap)
	}
	cols, cellSize := l.gridMetrics(width)
	rows := 0
	if len(objects) > 0 {
		rows = (len(objects) + cols - 1) / cols
	}
	height := float32(l.gap)
	if rows > 0 {
		height += float32(rows) * (cellSize.Height + float32(l.gap))
	}
	return fyne.NewSize(width, height)
}

func (l *ResponsiveGridLayout) setWidth(width float32) {
	l.lastWidth = width
}

func (l *ResponsiveGridLayout) gridMetrics(width float32) (int, fyne.Size) {
	gap := float32(l.gap)
	minimum := float32(l.minCellWidth)
	usableWidth := width - gap
	cols := int(usableWidth / (minimum + gap))
	if cols < 1 {
		cols = 1
	}

	cellWidth := (width - float32(cols+1)*gap) / float32(cols)
	if cellWidth < 1 {
		cellWidth = 1
	}
	cellHeight := cellWidth/float32(l.aspectRatio) + l.footerHeight
	return cols, fyne.NewSize(cellWidth, cellHeight)
}

// ResponsiveGrid owns a vertically scrolling custom layout. Supplying the
// viewport width before the scroller measures its content keeps its height and
// scrollbar correct throughout live window resizing.
type ResponsiveGrid struct {
	widget.BaseWidget

	layout   *ResponsiveGridLayout
	content  *fyne.Container
	scroller *container.Scroll
}

func NewResponsiveGrid(minCellWidth int, aspectRatio float64, gap int, footerHeight float32) *ResponsiveGrid {
	gridLayout := NewResponsiveGridLayoutWithFooter(minCellWidth, aspectRatio, gap, footerHeight)
	content := container.New(gridLayout)
	grid := &ResponsiveGrid{
		layout:   gridLayout,
		content:  content,
		scroller: container.NewVScroll(content),
	}
	grid.ExtendBaseWidget(grid)
	return grid
}

func (g *ResponsiveGrid) SetObjects(objects []fyne.CanvasObject) {
	g.content.Objects = objects
	g.content.Refresh()
	g.scroller.Refresh()
	g.Refresh()
}

func (g *ResponsiveGrid) ScrollToTop() {
	g.scroller.ScrollToTop()
}

func (g *ResponsiveGrid) CreateRenderer() fyne.WidgetRenderer {
	return &responsiveGridRenderer{
		grid:    g,
		objects: []fyne.CanvasObject{g.scroller},
	}
}

type responsiveGridRenderer struct {
	grid    *ResponsiveGrid
	objects []fyne.CanvasObject
}

func (r *responsiveGridRenderer) Layout(size fyne.Size) {
	r.grid.layout.setWidth(size.Width)
	r.grid.scroller.Resize(size)
	r.grid.scroller.Refresh()
}

func (r *responsiveGridRenderer) MinSize() fyne.Size {
	return fyne.NewSize(float32(r.grid.layout.minCellWidth+2*r.grid.layout.gap), 120)
}

func (r *responsiveGridRenderer) Refresh() {
	if width := r.grid.Size().Width; width > 0 {
		r.grid.layout.setWidth(width)
	}
	r.grid.scroller.Refresh()
}

func (r *responsiveGridRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *responsiveGridRenderer) Destroy() {}
