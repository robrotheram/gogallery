package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Sidebar owns one persistent inspector. Only its body scrolls, leaving the
// close button and editor actions reachable even in a short window.
type Sidebar struct {
	visible   bool
	title     string
	container *fyne.Container
	body      *fyne.Container
	heading   *widget.Label
	Content   fyne.CanvasObject
	OnToggle  func()
}

func NewSidebar(title string) *Sidebar { return &Sidebar{title: title} }

func NewTextEntry(text string, size float32) *canvas.Text {
	label := canvas.NewText(text, theme.Color(theme.ColorNameForeground))
	label.TextStyle = fyne.TextStyle{Bold: true}
	label.TextSize = size
	return label
}

func (s *Sidebar) Layout() fyne.CanvasObject {
	if s.container == nil {
		s.heading = widget.NewLabelWithStyle(s.title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		close := widget.NewButtonWithIcon("Close", theme.ContentClearIcon(), s.Hide)
		close.Importance = widget.LowImportance
		s.body = container.NewStack()
		header := container.NewBorder(nil, widget.NewSeparator(), nil, close, s.heading)
		panel := container.NewPadded(container.NewBorder(header, nil, nil, nil, s.body))
		s.container = container.NewStack(canvas.NewRectangle(theme.Color(theme.ColorNameMenuBackground)), panel)
		s.Refresh()
	}
	return s.container
}

func (s *Sidebar) SetContent(title string, content fyne.CanvasObject) {
	s.title, s.Content = title, content
	s.Show()
}
func (s *Sidebar) Refresh() {
	if s.container == nil {
		return
	}
	s.heading.SetText(s.title)
	if s.Content != nil {
		s.body.Objects = []fyne.CanvasObject{s.Content}
		s.body.Refresh()
	}
	if s.visible {
		s.container.Show()
	} else {
		s.container.Hide()
	}
}
func (s *Sidebar) Hide() {
	s.visible = false
	s.Refresh()
	if s.OnToggle != nil {
		s.OnToggle()
	}
}
func (s *Sidebar) Show() {
	s.visible = true
	s.Refresh()
	if s.OnToggle != nil {
		s.OnToggle()
	}
}

// NewInspectorWorkspace keeps the grid mounted. On narrow windows the
// inspector becomes a detail view; closing it restores the grid's position.
func NewInspectorWorkspace(body fyne.CanvasObject, sidebar *Sidebar) *fyne.Container {
	workspace := container.New(&inspectorLayout{sidebar: sidebar}, body, sidebar.Layout())
	sidebar.OnToggle = workspace.Refresh
	return workspace
}

type inspectorLayout struct{ sidebar *Sidebar }

func (l *inspectorLayout) MinSize(_ []fyne.CanvasObject) fyne.Size { return fyne.NewSize(360, 240) }
func (l *inspectorLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	body, panel := objects[0], objects[1]
	if !l.sidebar.visible {
		body.Show()
		panel.Hide()
		body.Move(fyne.Position{})
		body.Resize(size)
		return
	}
	panel.Show()
	if size.Width < 760 {
		body.Hide()
		panel.Move(fyne.Position{})
		panel.Resize(size)
		return
	}
	body.Show()
	const width float32 = 328
	body.Move(fyne.Position{})
	body.Resize(fyne.NewSize(size.Width-width-1, size.Height))
	panel.Move(fyne.NewPos(size.Width-width, 0))
	panel.Resize(fyne.NewSize(width, size.Height))
}
