package components

import (
	"fmt"
	"log"
	"net"
	"net/url"

	"gogallery/pkg/preview"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Header struct {
	Title       string
	onNavChange func(page string)
	server      *preview.Server
	buttons     map[string]*widget.Button
	compactNav  *widget.Select
	activePage  string
}

func NewHeader(title string, server *preview.Server, onNavChange func(page string)) *Header {
	if title == "" {
		title = "GoGallery"
	}
	return &Header{
		Title:       title,
		onNavChange: onNavChange,
		server:      server,
		buttons:     make(map[string]*widget.Button),
		activePage:  "Photos",
	}
}

func (h *Header) compactNavigation() *widget.Select {
	pages := []string{"Photos", "Albums", "Preview", "Tasks", "Settings"}
	h.compactNav = widget.NewSelect(pages, func(selected string) {
		if selected == "" || selected == h.activePage {
			return
		}
		if selected == "Preview" {
			h.Preview()
			h.compactNav.Selected = h.activePage
			h.compactNav.Refresh()
			return
		}
		if h.onNavChange != nil {
			h.onNavChange(selected)
		}
	})
	h.compactNav.Selected = h.activePage
	return h.compactNav
}

func (h *Header) navButton(page string, icon fyne.Resource) *widget.Button {
	button := widget.NewButtonWithIcon(page, icon, func() {
		if h.onNavChange != nil {
			h.onNavChange(page)
		}
	})
	button.Importance = widget.LowImportance
	h.buttons[page] = button
	return button
}

// Shell gives library navigation a stable home, switching to a compact
// toolbar only when the window cannot comfortably fit a navigation column.
func (h *Header) Shell(content fyne.CanvasObject) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(h.Title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.Truncation = fyne.TextTruncateEllipsis
	library := Hint("LIBRARY")
	photos := h.navButton("Photos", theme.MediaPhotoIcon())
	albums := h.navButton("Albums", theme.FolderIcon())
	tasks := h.navButton("Tasks", theme.HistoryIcon())
	settings := h.navButton("Settings", theme.SettingsIcon())
	for _, button := range h.buttons {
		button.Alignment = widget.ButtonAlignLeading
	}
	previewButton := widget.NewButtonWithIcon("Preview website", theme.VisibilityIcon(), h.Preview)
	previewButton.Importance = widget.LowImportance
	previewButton.Alignment = widget.ButtonAlignLeading
	top := container.NewVBox(title, widget.NewSeparator(), library, photos, albums)
	bottom := container.NewVBox(Hint("WORKSPACE"), tasks, settings, widget.NewSeparator(), previewButton)
	sidebar := container.NewStack(canvas.NewRectangle(theme.Color(theme.ColorNameMenuBackground)),
		container.NewPadded(container.NewBorder(top, bottom, nil, nil, layout.NewSpacer())))
	compact := container.NewPadded(container.NewBorder(nil, nil, widget.NewLabelWithStyle("GoGallery", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), h.compactNavigation()))
	return container.New(&shellLayout{}, sidebar, compact, content)
}

type shellLayout struct{}

func (*shellLayout) MinSize(_ []fyne.CanvasObject) fyne.Size { return fyne.NewSize(520, 400) }
func (*shellLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	sidebar, compact, content := objects[0], objects[1], objects[2]
	if size.Width < 900 {
		sidebar.Hide()
		compact.Show()
		height := compact.MinSize().Height
		compact.Move(fyne.Position{})
		compact.Resize(fyne.NewSize(size.Width, height))
		content.Move(fyne.NewPos(0, height))
		content.Resize(fyne.NewSize(size.Width, size.Height-height))
	} else {
		sidebar.Show()
		compact.Hide()
		const width float32 = 188
		sidebar.Move(fyne.Position{})
		sidebar.Resize(fyne.NewSize(width, size.Height))
		content.Move(fyne.NewPos(width, 0))
		content.Resize(fyne.NewSize(size.Width-width, size.Height))
	}
}

func (h *Header) SetActive(page string) {
	h.activePage = page
	if h.compactNav != nil {
		h.compactNav.Selected = page
		h.compactNav.Refresh()
	}
	for name, button := range h.buttons {
		if name == page {
			button.Importance = widget.HighImportance
		} else {
			button.Importance = widget.LowImportance
		}
		button.Refresh()
	}
}

func (h *Header) Preview() {
	if h.server == nil {
		return
	}
	status, _ := h.server.Status()
	if !status {
		if err := h.server.Start(); err != nil {
			showPreviewError(fmt.Errorf("could not start preview: %w", err))
			return
		}
	}
	host, port, err := net.SplitHostPort(h.server.Addr())
	if err != nil {
		showPreviewError(fmt.Errorf("could not read preview address: %w", err))
		return
	}
	if host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	u := &url.URL{Scheme: "http", Host: net.JoinHostPort(host, port)}
	app := fyne.CurrentApp()
	if app == nil {
		log.Print("Could not open preview: application unavailable")
		return
	}
	if err := app.OpenURL(u); err != nil {
		showPreviewError(fmt.Errorf("could not open preview in your browser: %w", err))
	}
}

func showPreviewError(err error) {
	log.Print(err)
	app := fyne.CurrentApp()
	if app != nil && len(app.Driver().AllWindows()) > 0 {
		dialog.ShowError(err, app.Driver().AllWindows()[0])
	}
}
