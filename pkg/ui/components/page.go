package components

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func Hint(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.Importance = widget.LowImportance
	label.Wrapping = fyne.TextWrapWord
	return label
}
func PageHeading(title, description string) fyne.CanvasObject {
	return container.NewVBox(NewTextEntry(title, 24), Hint(description))
}

// WorkspaceFrame keeps utility pages aligned beside navigation while leaving
// comfortable margins and limiting line lengths on fullscreen displays.
func WorkspaceFrame(content fyne.CanvasObject, width float32) fyne.CanvasObject {
	return container.New(&workspaceFrameLayout{width: width}, content)
}

type workspaceFrameLayout struct{ width float32 }

func (*workspaceFrameLayout) MinSize([]fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(360, 300)
}
func (l *workspaceFrameLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	inset := float32(32)
	if size.Width < 760 {
		inset = 16
	}
	width := max(0, min(l.width, size.Width-2*inset))
	for _, object := range objects {
		object.Move(fyne.NewPos(inset, inset))
		object.Resize(fyne.NewSize(width, max(0, size.Height-2*inset)))
	}
}

// ReadingWidth prevents forms stretching across an entire desktop.
func ReadingWidth(content fyne.CanvasObject, width float32) fyne.CanvasObject {
	return container.New(&readingLayout{width: width}, content)
}

type readingLayout struct{ width float32 }

func (l *readingLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(320, objects[0].MinSize().Height)
}
func (l *readingLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	w := min(l.width, size.Width)
	objects[0].Move(fyne.NewPos((size.Width-w)/2, 0))
	objects[0].Resize(fyne.NewSize(w, max(size.Height, objects[0].MinSize().Height)))
}
func Field(title string, control fyne.CanvasObject) fyne.CanvasObject {
	return container.NewVBox(widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), control)
}

func PhotoCount(count int) string {
	if count == 1 {
		return "1 photo"
	}
	return fmt.Sprintf("%d photos", count)
}
