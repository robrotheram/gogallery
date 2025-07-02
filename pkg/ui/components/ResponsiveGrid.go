package components

import "fyne.io/fyne/v2"

// --- Responsive grid layout ---
type ResponsiveGridLayout struct {
	minCellWidth int
	aspectRatio  float64
	gap          int
	lastWidth    float32 // <-- add this
}

func (l *ResponsiveGridLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	l.lastWidth = size.Width
	w := float64(size.Width)
	minCellWidth := float64(l.minCellWidth)
	gap := float64(l.gap)
	cols := max(int(w/(minCellWidth+gap)), 1)
	cellW := (w - float64(cols+1)*gap) / float64(cols)
	cellH := cellW / l.aspectRatio

	for i, obj := range objects {
		row := i / cols
		col := i % cols
		x := float32(gap + float64(col)*(cellW+gap))
		y := float32(gap + float64(row)*(cellH+gap))
		obj.Resize(fyne.NewSize(float32(cellW), float32(cellH)))
		obj.Move(fyne.NewPos(x, y))
	}
}
func (l *ResponsiveGridLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minCellWidth := float32(l.minCellWidth)
	gap := float32(l.gap)
	n := len(objects)
	if n == 0 {
		cellH := minCellWidth / float32(l.aspectRatio)
		return fyne.NewSize(minCellWidth+2*gap, cellH+2*gap)
	}
	w := minCellWidth + 2*float32(l.gap)
	// Use lastWidth for dynamic calculation, fallback to 1 column if not set
	width := l.lastWidth
	if width < minCellWidth+2*gap {
		width = minCellWidth + 2*gap
	}
	cols := int(width / (minCellWidth + gap))
	if cols < 1 {
		cols = 1
	}
	// Calculate dynamic cell width and height
	cellW := (float64(width) - float64(cols+1)*float64(gap)) / float64(cols)
	cellH := float32(cellW / l.aspectRatio)

	rows := (n + cols - 1) / cols
	h := float32(rows)*(cellH+gap) + gap

	return fyne.NewSize(w, h)
}
