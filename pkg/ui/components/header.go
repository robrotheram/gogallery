package components

import (
	"log"
	"net"
	"net/url"

	"gogallery/pkg/preview"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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
	layout      *fyne.Container
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
		activePage:  "Gallery",
	}
}

func (h *Header) nav() *fyne.Container {
	gallery := h.navButton("Gallery", theme.GridIcon())
	collection := h.navButton("Collections", theme.FolderIcon())
	preview := widget.NewButtonWithIcon("Preview", theme.VisibilityIcon(), func() {
		h.Preview()
	})
	preview.Importance = widget.MediumImportance
	tasks := h.navButton("Tasks", theme.ContentPasteIcon())
	settings := h.navButton("Settings", theme.SettingsIcon())
	return container.NewHBox(gallery, collection, preview, tasks, settings)
}

func (h *Header) compactNavigation() *widget.Select {
	pages := []string{"Gallery", "Collections", "Preview", "Tasks", "Settings"}
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

func (h *Header) Layout() *fyne.Container {
	if h.layout != nil {
		return h.layout
	}

	clickableTitle := NewClickableTitle(h.Title, func() {
		if h.onNavChange != nil {
			h.onNavChange("Gallery")
		}
	})
	fullNavigation := h.nav()
	compactNavigation := h.compactNavigation()
	headerBox := container.New(&responsiveHeaderLayout{gap: 16}, clickableTitle, fullNavigation, compactNavigation)

	h.layout = container.NewVBox(
		container.NewPadded(headerBox),
		widget.NewSeparator(),
	)
	return h.layout
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
	status, _ := h.server.Status()
	if !status {
		if err := h.server.Start(); err != nil {
			log.Printf("Could not start preview server: %v", err)
			return
		}
	}
	host, port, err := net.SplitHostPort(h.server.Addr())
	if err != nil {
		log.Printf("Could not parse preview address: %v", err)
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
		log.Printf("Could not open preview URL %s: %v", u, err)
	}
}

// responsiveHeaderLayout keeps the full labelled navigation when it fits and
// swaps to a compact page selector before controls can overlap or clip.
type responsiveHeaderLayout struct {
	gap float32
}

func (l *responsiveHeaderLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 3 {
		return
	}
	brand, fullNavigation, compactNavigation := objects[0], objects[1], objects[2]
	useFullNavigation := brand.MinSize().Width+fullNavigation.MinSize().Width+l.gap <= size.Width

	navigation := compactNavigation
	if useFullNavigation {
		fullNavigation.Show()
		compactNavigation.Hide()
		navigation = fullNavigation
	} else {
		fullNavigation.Hide()
		compactNavigation.Show()
	}

	brandSize := brand.MinSize()
	navigationSize := navigation.MinSize()
	brand.Resize(brandSize)
	navigation.Resize(navigationSize)
	brand.Move(fyne.NewPos(0, (size.Height-brandSize.Height)/2))
	navigation.Move(fyne.NewPos(size.Width-navigationSize.Width, (size.Height-navigationSize.Height)/2))
}

func (l *responsiveHeaderLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) < 3 {
		return fyne.Size{}
	}
	brandSize := objects[0].MinSize()
	fullSize := objects[1].MinSize()
	compactSize := objects[2].MinSize()
	return fyne.NewSize(
		brandSize.Width+compactSize.Width+l.gap,
		max(brandSize.Height, max(fullSize.Height, compactSize.Height)),
	)
}
