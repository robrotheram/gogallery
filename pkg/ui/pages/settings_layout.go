package pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gogallery/pkg/ui/components"
)

// Give settings comfortable inputs without increasing padding on every label.
type settingsTheme struct{ fyne.Theme }

func (t settingsTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameInnerPadding:
		return 12
	case theme.SizeNameText:
		return 15
	default:
		return t.Theme.Size(name)
	}
}

type settingsGroup struct {
	title string
	items []*widget.FormItem
}

func (g settingsGroup) content() fyne.CanvasObject {
	heading := container.New(layout.NewCustomPaddedLayout(0, 0, 8, 0), components.NewTextEntry(g.title, 18))
	fields := []fyne.CanvasObject{heading}
	for _, item := range g.items {
		control := container.NewThemeOverride(item.Widget, settingsTheme{fyne.CurrentApp().Settings().Theme()})
		field := container.NewVBox(control)
		if item.Text != "" {
			field = container.NewVBox(components.Field(item.Text, control))
		}
		if item.HintText != "" {
			field.Add(components.Hint(item.HintText))
		}
		fields = append(fields, field)
	}
	body := container.New(layout.NewCustomPaddedVBoxLayout(16), fields...)
	return widget.NewCard("", "", container.New(layout.NewCustomPaddedLayout(8, 20, 12, 12), body))
}

// Prepare column widths before the scroller measures height. This also lets
// wrapped hints reflow when moving between a large monitor and a small window.
type settingsBody struct {
	widget.BaseWidget
	content *fyne.Container
	scroll  *container.Scroll
	columns *settingsColumnsLayout
}

func newSettingsBody(introduction fyne.CanvasObject, groups ...fyne.CanvasObject) *settingsBody {
	columns := &settingsColumnsLayout{}
	body := &settingsBody{columns: columns}
	body.content = container.New(columns, append([]fyne.CanvasObject{introduction}, groups...)...)
	body.scroll = container.NewVScroll(body.content)
	body.ExtendBaseWidget(body)
	return body
}

func (b *settingsBody) CreateRenderer() fyne.WidgetRenderer { return &settingsBodyRenderer{body: b} }

type settingsBodyRenderer struct{ body *settingsBody }

func (r *settingsBodyRenderer) MinSize() fyne.Size { return fyne.NewSize(280, 80) }
func (r *settingsBodyRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.body.scroll}
}
func (r *settingsBodyRenderer) Destroy() {}
func (r *settingsBodyRenderer) Layout(size fyne.Size) {
	// The tab and group titles identify the section when a short viewport
	// cannot comfortably fit an introduction as well as its first field.
	if size.Height < 240 {
		r.body.content.Objects[0].Hide()
	} else {
		r.body.content.Objects[0].Show()
	}
	r.body.columns.Layout(r.body.content.Objects, size)
	r.body.scroll.Resize(size)
	r.body.scroll.Refresh()
}
func (r *settingsBodyRenderer) Refresh() { r.Layout(r.body.Size()) }

type settingsColumnsLayout struct{ width float32 }

func (l *settingsColumnsLayout) columnCount(objects []fyne.CanvasObject) int {
	if l.width >= 960 && len(objects) > 2 {
		return 2
	}
	return 1
}

func (l *settingsColumnsLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	l.width = size.Width
	const gutter float32 = 24
	// Leave room at the right for the scrollbar and around card shadows.
	width := max(0, size.Width-12)
	resize := func(object fyne.CanvasObject, width float32) float32 {
		object.Resize(fyne.NewSize(width, object.MinSize().Height))
		height := object.MinSize().Height
		object.Resize(fyne.NewSize(width, height))
		return height
	}
	intro := objects[0]
	intro.Move(fyne.NewPos(0, 16))
	y := float32(16)
	if intro.Visible() {
		y += resize(intro, width) + gutter
	}
	columns := l.columnCount(objects)
	cellWidth := (width - float32(columns-1)*gutter) / float32(columns)
	if len(objects) == 2 {
		cellWidth = min(cellWidth, 760)
	}
	heights := make([]float32, columns)
	for column := range heights {
		heights[column] = y
	}
	for _, object := range objects[1:] {
		column := shortestSettingsColumn(heights)
		object.Move(fyne.NewPos(float32(column)*(cellWidth+gutter), heights[column]))
		heights[column] += resize(object, cellWidth) + gutter
	}
}

func (l *settingsColumnsLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	height := float32(16)
	if objects[0].Visible() {
		height += objects[0].MinSize().Height + 24
	}
	columns := l.columnCount(objects)
	heights := make([]float32, columns)
	for column := range heights {
		heights[column] = height
	}
	for _, object := range objects[1:] {
		column := shortestSettingsColumn(heights)
		heights[column] += object.MinSize().Height + 24
	}
	for _, columnHeight := range heights {
		height = max(height, columnHeight)
	}
	return fyne.NewSize(280, height)
}

// Stack shorter sections together so Preview and AI use the right column
// without a large blank row beside Appearance. Narrow windows retain order.
func shortestSettingsColumn(heights []float32) int {
	column := 0
	for i := 1; i < len(heights); i++ {
		if heights[i] < heights[column] {
			column = i
		}
	}
	return column
}
