package pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gogallery/pkg/monitor"
)

type taskActionTheme struct{ fyne.Theme }

func (t taskActionTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 15
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameInnerPadding:
		return 10
	default:
		return t.Theme.Size(name)
	}
}

// Set the grid's width before the scroller measures its height, so activity
// moves down immediately when the task buttons wrap into more rows.
type tasksPageLayout struct{ buttons *taskButtonsLayout }

func (*tasksPageLayout) MinSize([]fyne.CanvasObject) fyne.Size { return fyne.NewSize(320, 240) }
func (l *tasksPageLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	heading, scroll := objects[0], objects[1].(*container.Scroll)
	heading.Resize(fyne.NewSize(size.Width, heading.MinSize().Height))
	height := heading.MinSize().Height
	heading.Resize(fyne.NewSize(size.Width, height))
	l.buttons.width = max(0, size.Width-12)
	scroll.Move(fyne.NewPos(0, height+20))
	scroll.Resize(fyne.NewSize(size.Width, max(0, size.Height-height-20)))
	scroll.Refresh()
}

type taskButtonsLayout struct{ width float32 }

const (
	taskButtonHeight float32 = 56
	taskButtonGap    float32 = 12
)

func (l *taskButtonsLayout) columns() int {
	if l.width >= 960 {
		return 4
	}
	if l.width >= 460 {
		return 2
	}
	return 1
}
func (l *taskButtonsLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	rows := (len(objects) + l.columns() - 1) / l.columns()
	return fyne.NewSize(240, float32(rows)*taskButtonHeight+float32(max(0, rows-1))*taskButtonGap)
}
func (l *taskButtonsLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	l.width = size.Width
	columns := l.columns()
	width := max(0, (size.Width-float32(columns-1)*taskButtonGap)/float32(columns))
	for i, object := range objects {
		object.Move(fyne.NewPos(float32(i%columns)*(width+taskButtonGap), float32(i/columns)*(taskButtonHeight+taskButtonGap)))
		object.Resize(fyne.NewSize(width, taskButtonHeight))
	}
}

func taskStatusBadge(text string, state monitor.ProcessState) fyne.CanvasObject {
	label := widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	icon := theme.HistoryIcon()
	switch state {
	case monitor.COMPLETE:
		label.Importance = widget.SuccessImportance
		icon = theme.NewSuccessThemedResource(theme.ConfirmIcon())
	case monitor.ERROR:
		label.Importance = widget.DangerImportance
		icon = theme.NewErrorThemedResource(theme.ErrorIcon())
	case monitor.RUNNING:
		icon = theme.ViewRefreshIcon()
	case monitor.STOPPED:
		icon = theme.MediaStopIcon()
	}
	background := canvas.NewRectangle(theme.Color(theme.ColorNameButton))
	background.CornerRadius = 6
	return container.NewStack(background, container.New(layout.NewCustomPaddedLayout(0, 0, 8, 4),
		container.NewHBox(widget.NewIcon(icon), label)))
}
