package pages

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gogallery/pkg/ui/components"
)

func TestTaskButtonsWrapWithoutOverlappingActivity(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	grid := &taskButtonsLayout{}
	page := &TasksPage{}
	actions := container.New(grid)
	for _, name := range []string{"Scan library", "Build website", "Start preview", "Publish to Netlify"} {
		actions.Add(page.action(name, theme.MediaPhotoIcon(), func() {}))
	}
	activity := widget.NewLabel("Activity")
	body := container.New(layout.NewCustomPaddedVBoxLayout(16), actions, activity)
	scroll := container.NewVScroll(container.New(layout.NewCustomPaddedLayout(0, 12, 0, 12), body))
	content := container.New(&tasksPageLayout{buttons: grid}, widget.NewLabel("Tasks"), scroll)
	window := test.NewWindow(components.WorkspaceFrame(content, 1120))
	defer window.Close()
	window.SetPadded(false)
	for _, size := range []fyne.Size{fyne.NewSize(1280, 800), fyne.NewSize(600, 640), fyne.NewSize(520, 400), fyne.NewSize(360, 640), fyne.NewSize(1280, 800)} {
		window.Resize(size)
		expectedColumns := 4
		if size.Width <= 600 {
			expectedColumns = 2
		}
		if size.Width < 520 {
			expectedColumns = 1
		}
		for i, button := range page.actions {
			pos := actions.Objects[i].Position()
			if (i < expectedColumns) != (pos.Y == 0) {
				t.Fatalf("incorrect wrapping at %v: button %d at %v", size, i, pos)
			}
			if button.Size().Width < button.MinSize().Width || button.Size().Height < button.MinSize().Height || button.Size().Height > 64 {
				t.Fatalf("button is clipped or oversized at %v: %v, minimum %v", size, button.Size(), button.MinSize())
			}
			absolute := app.Driver().AbsolutePositionForObject(button)
			if absolute.X < 0 || absolute.X+button.Size().Width > size.Width {
				t.Fatalf("button extends past the window at %v: %v", size, absolute)
			}
		}
		last := actions.Objects[len(actions.Objects)-1]
		if activity.Position().Y < actions.Position().Y+last.Position().Y+last.Size().Height {
			t.Fatalf("activity overlaps the wrapped buttons at %v", size)
		}
		scroll.ScrollToBottom()
		position := app.Driver().AbsolutePositionForObject(activity)
		if position.Y+activity.Size().Height > size.Height {
			t.Fatalf("activity is unreachable at %v", size)
		}
		scroll.ScrollToTop()
	}
}
