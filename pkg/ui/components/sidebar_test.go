package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"testing"
)

func TestInspectorKeepsItsWidthAndRestoresGrid(t *testing.T) {
	app := test.NewApp()
	t.Cleanup(app.Quit)
	body := sizedRectangle(300, 200)
	sidebar := NewSidebar("Details")
	workspace := NewInspectorWorkspace(body, sidebar)
	window := test.NewWindow(workspace)
	defer window.Close()
	window.SetPadded(false)
	window.Resize(fyne.NewSize(1100, 600))
	sidebar.SetContent("Photo details", widget.NewLabel("A long unbroken filename that must not change the inspector width"))
	if got := sidebar.Layout().Size().Width; got != 328 {
		t.Fatalf("inspector width = %v", got)
	}
	if body.Size().Width != 771 {
		t.Fatalf("grid width = %v", body.Size().Width)
	}
	panel := sidebar.Layout()
	window.Resize(fyne.NewSize(600, 500))
	if body.Visible() || !panel.Visible() || panel.Size().Width != 600 {
		t.Fatal("narrow inspector should fill the content area")
	}
	sidebar.Hide()
	if !body.Visible() || panel.Visible() || body.Size().Width != 600 {
		t.Fatal("closing inspector did not restore the grid")
	}
	sidebar.Show()
	window.Resize(fyne.NewSize(1100, 600))
	if !body.Visible() || panel != sidebar.Layout() || panel.Size().Width != 328 {
		t.Fatal("resizing replaced or broke the inspector")
	}
}
