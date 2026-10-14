package components

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
	"gogallery/pkg/datastore"
	"image"
	"path/filepath"
)

// PhotoPreview contains the image inside the space it is given. Portraits and
// panoramas never impose their intrinsic dimensions on the surrounding form.
func PhotoPreview(db *datastore.DataStore, pic datastore.Picture, original bool) fyne.CanvasObject {
	background := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	background.SetMinSize(fyne.NewSize(0, 190))
	label := widget.NewLabel("Loading preview…")
	label.Alignment = fyne.TextAlignCenter
	stack := container.NewStack(background, container.NewCenter(label))
	if pic.Id == "" {
		label.SetText("No cover selected")
		return stack
	}
	go func() {
		var photo image.Image
		var err error
		if original {
			photo, err = pic.Load()
			if err == nil {
				photo = imaging.Fit(photo, 2400, 1800, imaging.Lanczos)
			}
		} else {
			photo, err = (&Image{dataStore: db}).loadImage(pic)
		}
		fyne.Do(func() {
			if err != nil {
				label.SetText("Preview unavailable")
				return
			}
			img := canvas.NewImageFromImage(photo)
			img.FillMode = canvas.ImageFillContain
			stack.Objects = []fyne.CanvasObject{background, img}
			stack.Refresh()
		})
	}()
	return stack
}

func ShowPhotoViewer(db *datastore.DataStore, pictures []datastore.Picture, selected string) {
	if len(pictures) == 0 || fyne.CurrentApp() == nil {
		return
	}
	index := 0
	for i, pic := range pictures {
		if pic.Id == selected {
			index = i
			break
		}
	}
	window := fyne.CurrentApp().NewWindow("Photo — GoGallery")
	stage := container.NewStack()
	title := widget.NewLabel("")
	title.Truncation = fyne.TextTruncateEllipsis
	count := widget.NewLabel("")
	previous := widget.NewButtonWithIcon("Previous", theme.NavigateBackIcon(), nil)
	next := widget.NewButtonWithIcon("Next", theme.NavigateNextIcon(), nil)
	show := func() {
		pic := pictures[index]
		name := pic.Name
		if name == "" {
			name = filepath.Base(pic.Path)
		}
		title.SetText(name)
		count.SetText(fmt.Sprintf("%d of %d", index+1, len(pictures)))
		window.SetTitle(name + " — GoGallery")
		stage.Objects = []fyne.CanvasObject{PhotoPreview(db, pic, true)}
		stage.Refresh()
		previous.Enable()
		next.Enable()
		if index == 0 {
			previous.Disable()
		}
		if index == len(pictures)-1 {
			next.Disable()
		}
	}
	previous.OnTapped = func() {
		if index > 0 {
			index--
			show()
		}
	}
	next.OnTapped = func() {
		if index+1 < len(pictures) {
			index++
			show()
		}
	}
	close := widget.NewButtonWithIcon("Close", theme.ContentClearIcon(), window.Close)
	toolbar := container.NewBorder(nil, nil, nil, close, title)
	footer := container.NewHBox(layout.NewSpacer(), previous, count, next, layout.NewSpacer())
	window.SetContent(container.NewBorder(container.NewPadded(toolbar), container.NewPadded(footer), nil, nil, stage))
	window.Canvas().SetOnTypedKey(func(event *fyne.KeyEvent) {
		switch event.Name {
		case fyne.KeyLeft:
			previous.OnTapped()
		case fyne.KeyRight:
			next.OnTapped()
		case fyne.KeyEscape:
			window.Close()
		}
	})
	show()
	window.Resize(fyne.NewSize(1000, 740))
	window.Show()
}
