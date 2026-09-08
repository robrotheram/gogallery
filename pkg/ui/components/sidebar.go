package components

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Sidebar struct {
	visible   bool
	title     string
	container fyne.CanvasObject // Reference to the sidebar container for refresh
	Content   fyne.CanvasObject // Reference to the sidebar content for refresh
	OnToggle  func()            // Callback for when visibility changes
}

func NewSidebar(title string) *Sidebar {
	return &Sidebar{
		visible:  false,
		title:    title,
		Content:  nil, // Will be set in Layout
		OnToggle: nil, // Can be set later if needed
	}
}

func NewTextEntry(textStr string, size float32) *canvas.Text {
	text := canvas.NewText(textStr, nil)
	text.TextStyle = fyne.TextStyle{Bold: true}
	text.TextSize = size // Larger font size
	return text
}

func (s *Sidebar) Layout() fyne.CanvasObject {
	// If not visible, return an empty container with zero size
	if !s.visible {
		emptyContainer := container.NewWithoutLayout()
		emptyContainer.Resize(fyne.NewSize(0, 0))
		s.container = emptyContainer
		return s.container
	}

	// Close button
	closeBtn := widget.NewButtonWithIcon("", theme.ContentClearIcon(), s.Hide)
	closeBtn.Importance = widget.LowImportance
	closeBtn.Alignment = widget.ButtonAlignTrailing
	closeBtnBox := container.NewVBox(
		layout.NewSpacer(),
		closeBtn,
		layout.NewSpacer(),
	)

	titleRow := container.NewHBox(
		NewTextEntry(s.title, 22),
		layout.NewSpacer(),
		closeBtnBox,
	)
	scrollContent := container.NewBorder(
		titleRow,  // top
		nil,       // bottom
		nil,       // left
		nil,       // right
		s.Content, // main content
	)
	panelBackground := canvas.NewRectangle(theme.Color(theme.ColorNameMenuBackground))
	panel := container.NewStack(panelBackground, container.NewPadded(scrollContent))
	s.container = container.NewBorder(nil, nil, widget.NewSeparator(), nil, container.NewVScroll(panel))
	return s.container
}

// Refresh the sidebar UI (call this after ShowImage)
func (s *Sidebar) Refresh() {
	log.Println("Refreshing sidebar")
	if s.visible {
		s.container.Show()
	} else {
		s.container.Hide()
	}
}

func (s *Sidebar) Hide() {
	s.visible = false
	if s.OnToggle != nil {
		s.OnToggle()
	}
}

func (s *Sidebar) Show() {
	s.visible = true
	if s.OnToggle != nil {
		s.OnToggle()
	}
}
