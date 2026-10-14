package pages

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestSettingsGroupsReflowAndRemainScrollable(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	first := widget.NewEntry()
	last := widget.NewEntry()
	profile := (settingsGroup{"Profile", []*widget.FormItem{
		{Text: "Photographer", Widget: first, HintText: strings.Repeat("A hint that wraps as the window narrows. ", 8)},
	}}).content()
	links := (settingsGroup{"Links", []*widget.FormItem{{Text: "Website", Widget: last}}}).content()
	body := newSettingsBody(widget.NewLabel("About the photographer"), profile, links)
	window := test.NewWindow(body)
	defer window.Close()
	window.SetPadded(false)
	for _, width := range []float32{1240, 520, 960, 600, 1240} {
		window.Resize(fyne.NewSize(width, 300))
		body.scroll.ScrollToTop()
		if width >= 960 {
			if links.Position().Y != profile.Position().Y || links.Position().X < profile.Position().X+profile.Size().Width+20 {
				t.Fatalf("groups are not side by side at width %v", width)
			}
		} else if links.Position().X != profile.Position().X || links.Position().Y < profile.Position().Y+profile.Size().Height+20 {
			t.Fatalf("groups overlap instead of stacking at width %v", width)
		}
		for _, field := range []*widget.Entry{first, last} {
			position := app.Driver().AbsolutePositionForObject(field)
			if position.X < 0 || position.X+field.Size().Width > width {
				t.Fatalf("field extends outside window at width %v: %v, %v", width, position, field.Size())
			}
			if field.Size().Height < 42 {
				t.Fatalf("settings input is too cramped: %v", field.Size())
			}
		}
		body.scroll.ScrollToBottom()
		position := app.Driver().AbsolutePositionForObject(last)
		if position.Y+last.Size().Height > 300 {
			t.Fatalf("last field cannot be reached at width %v: %v", width, position)
		}
	}
}

func TestSettingsFrameKeepsControlsInsideSmallAndLargeWindows(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	page := NewSettingsPage(nil).Layout()
	window := test.NewWindow(page)
	defer window.Close()
	window.SetPadded(false)
	var tabs *container.AppTabs
	var inspect func(fyne.CanvasObject, fyne.Size)
	inspect = func(object fyne.CanvasObject, size fyne.Size) {
		if !object.Visible() {
			return
		}
		if tabBar, ok := object.(*container.AppTabs); ok {
			tabs = tabBar
		}
		if button, ok := object.(*widget.Button); ok && (button.Text == "Save changes" || button.Text == "Reset changes") {
			pos := app.Driver().AbsolutePositionForObject(button)
			if pos.X < 0 || pos.Y < 0 || pos.X+button.Size().Width > size.Width || pos.Y+button.Size().Height > size.Height {
				t.Fatalf("%s is outside %v: %v, %v", button.Text, size, pos, button.Size())
			}
		}
		if label, ok := object.(*canvas.Text); ok && label.Text == "Settings" {
			pos := app.Driver().AbsolutePositionForObject(label)
			if pos.X < 16 || pos.X > 40 || pos.Y < 16 {
				t.Fatalf("settings header lost its navigation margin at %v: %v", size, pos)
			}
		}
		if c, ok := object.(*fyne.Container); ok {
			for _, child := range c.Objects {
				inspect(child, size)
			}
		} else if w, ok := object.(fyne.Widget); ok {
			for _, child := range test.WidgetRenderer(w).Objects() {
				inspect(child, size)
			}
		}
	}
	for _, size := range []fyne.Size{fyne.NewSize(520, 400), fyne.NewSize(600, 640), fyne.NewSize(1920, 1080), fyne.NewSize(2560, 1440), fyne.NewSize(520, 400)} {
		window.Resize(size)
		inspect(page, size)
		if tabs == nil {
			t.Fatal("settings tabs missing")
		}
		for i := range tabs.Items {
			tabs.SelectIndex(i)
			inspect(page, size)
		}
	}
}

func TestPreviewAndAIStackBesideAppearance(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	appearance := widget.NewMultiLineEntry()
	appearance.SetMinRowsVisible(15)
	preview := widget.NewLabel("Preview")
	ai := widget.NewLabel("AI")
	body := newSettingsBody(widget.NewLabel("Application"), appearance, preview, ai)
	window := test.NewWindow(body)
	defer window.Close()
	window.SetPadded(false)
	for _, width := range []float32{1400, 520, 1400} {
		window.Resize(fyne.NewSize(width, 600))
		if width >= 960 {
			if preview.Position().X <= appearance.Position().X || ai.Position().X != preview.Position().X {
				t.Fatal("Preview and AI must share the right column")
			}
			if ai.Position().Y < preview.Position().Y+preview.Size().Height || ai.Position().Y >= appearance.Position().Y+appearance.Size().Height {
				t.Fatal("AI must follow Preview without waiting for the taller Appearance section")
			}
		} else if preview.Position().Y < appearance.Position().Y+appearance.Size().Height || ai.Position().Y < preview.Position().Y+preview.Size().Height {
			t.Fatal("narrow settings sections must stack without overlap")
		}
	}
}
