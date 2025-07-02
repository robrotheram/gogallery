package components

import (
	"fmt"
	"testingFyne/pkg/datastore"
	"testingFyne/pkg/preview"

	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Header struct {
	Title       string
	onNavChange func(page string)
	server      *preview.Server
}

func NewHeader(title string, db *datastore.DataStore, server *preview.Server, onNavChange func(page string)) *Header {
	return &Header{
		Title:       title,
		onNavChange: onNavChange,
		server:      server,
	}
}

func (h *Header) nav() *fyne.Container {
	preview := widget.NewButtonWithIcon("Preview", theme.VisibilityIcon(), func() {
		h.Preview()
	})
	tasks := widget.NewButtonWithIcon("Tasks", theme.ContentPasteIcon(), func() {
		h.onNavChange("Tasks")
	})
	settings := widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), func() {
		h.onNavChange("Settings")
	})
	navButtons := []fyne.CanvasObject{preview, tasks, settings}
	return container.NewHBox(container.NewHBox(navButtons...))
}

func (h *Header) Layout() *fyne.Container {
	// Clickable title that looks like a real title
	clickableTitle := NewClickableTitle(h.Title, func() {
		h.onNavChange("Gallery")
	})
	leftPad := canvas.NewRectangle(nil)
	leftPad.SetMinSize(fyne.NewSize(12, 0))
	headerBox := container.NewHBox(
		leftPad,
		clickableTitle,
		leftPad,
		layout.NewSpacer(),
		h.nav(),
		leftPad,
	)
	headerBox.Resize(fyne.NewSize(0, 64))

	return container.NewVBox(
		headerBox,
		widget.NewSeparator(),
	)
}

func (h *Header) Preview() {
	status, _ := h.server.Status()
	if !status {
		h.server.Start()
	}
	u, _ := url.Parse(fmt.Sprintf("http://%s", h.server.Addr()))
	fyne.CurrentApp().OpenURL(u)
}

/*
	if alb, err := h.Albums.GetAll(); err == nil {
		options := make([]string, len(alb))
		for i, a := range alb {
			options[i] = a.Name
		}
		h.FilterList.Options = options
	} else {
		h.FilterList.Options = []string{"No Albums Found"}
	}
	h.FilterList.PlaceHolder = "Select Album"
	h.FilterList.Selected = h.Title // Set the initial selected album to the title

*/
