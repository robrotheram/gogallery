package components

import (
	"bytes"
	"fmt"
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"gogallery/pkg/pipeline"
	"image"
	"image/color"
	"io"
	"log"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// The height reserves a 3:2 photo area plus the caption footer.
var imageTileMinSize = fyne.NewSize(260, 218)

// Image is a reusable image tile with asynchronous thumbnail loading.
type Image struct {
	widget.BaseWidget

	dataStore *datastore.DataStore
	pic       datastore.Picture

	content       *fyne.Container
	placeholderUI *fyne.Container
	imageObj      *canvas.Image
	hoverBorder   *canvas.Rectangle
	activeBorder  *canvas.Rectangle
	captionLabel  *widget.Label
	showCaption   bool
	isLoading     bool
	isHovered     bool
	isActive      bool
	onLoad        func()
	onClick       func(datastore.Picture)
}

func NewImage(dataStore *datastore.DataStore, pic datastore.Picture, onLoad func(), onClick func(datastore.Picture)) *Image {
	return NewImageWithAspect(dataStore, pic, onLoad, onClick, 6, 4)
}

// NewImageWithAspect is kept for compatibility. ImageFillCover now performs the
// display crop, avoiding a decode/crop/encode/decode cycle for every thumbnail.
func NewImageWithAspect(dataStore *datastore.DataStore, pic datastore.Picture, onLoad func(), onClick func(datastore.Picture), _, _ int) *Image {
	img := &Image{
		dataStore: dataStore,
		pic:       pic,
		onLoad:    onLoad,
		onClick:   onClick,
	}
	img.ExtendBaseWidget(img)
	img.createBorders()
	img.createPlaceholderUI()
	img.updateCaption(pic)

	if pic.Id != "" {
		img.loadImageAsync()
	}

	return img
}

// Tapped implements fyne.Tappable.
func (img *Image) Tapped(_ *fyne.PointEvent) {
	if img.onClick != nil && img.pic.Id != "" {
		img.onClick(img.pic)
	}
}

// MouseIn implements desktop.Hoverable.
func (img *Image) MouseIn(_ *desktop.MouseEvent) {
	img.isHovered = true
	img.updateBorders()
}

// MouseOut implements desktop.Hoverable.
func (img *Image) MouseOut() {
	img.isHovered = false
	img.updateBorders()
}

// MouseMoved implements desktop.Hoverable.
func (img *Image) MouseMoved(_ *desktop.MouseEvent) {}

// Cursor returns the cursor to show when hovering.
func (img *Image) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

func (img *Image) createBorders() {
	img.hoverBorder = canvas.NewRectangle(color.Transparent)
	img.hoverBorder.StrokeWidth = 2
	img.hoverBorder.StrokeColor = theme.Color(theme.ColorNamePrimary)
	img.hoverBorder.Hide()

	img.activeBorder = canvas.NewRectangle(color.Transparent)
	img.activeBorder.StrokeWidth = 3
	img.activeBorder.StrokeColor = theme.Color(theme.ColorNameFocus)
	img.activeBorder.Hide()
}

func (img *Image) updateBorders() {
	if img.isActive {
		img.hoverBorder.Hide()
		img.activeBorder.Show()
	} else if img.isHovered {
		img.activeBorder.Hide()
		img.hoverBorder.Show()
	} else {
		img.hoverBorder.Hide()
		img.activeBorder.Hide()
	}
}

func (img *Image) CreateRenderer() fyne.WidgetRenderer {
	if img.content == nil {
		img.content = img.placeholderUI
	}
	return &imageRenderer{image: img}
}

type imageRenderer struct {
	image *Image
}

func (r *imageRenderer) Layout(size fyne.Size) {
	if r.image.content != nil {
		r.image.content.Resize(size)
	}
	r.image.hoverBorder.Resize(size)
	r.image.activeBorder.Resize(size)
}

func (r *imageRenderer) MinSize() fyne.Size {
	return imageTileMinSize
}

func (r *imageRenderer) Refresh() {
	if r.image.content != nil {
		r.image.content.Refresh()
	}
}

func (r *imageRenderer) Objects() []fyne.CanvasObject {
	objects := make([]fyne.CanvasObject, 0, 3)
	if r.image.content != nil {
		objects = append(objects, r.image.content)
	}
	return append(objects, r.image.hoverBorder, r.image.activeBorder)
}

func (r *imageRenderer) Destroy() {}

// SetPicture reuses this tile for another picture. This is used by GridWrap so
// only the visible set of image widgets needs to exist at any time.
func (img *Image) SetPicture(pic datastore.Picture) {
	if img.pic.Id == pic.Id && (img.imageObj != nil || img.isLoading) {
		return
	}

	img.pic = pic
	img.imageObj = nil
	img.updateCaption(pic)
	img.showPlaceholder()

	if pic.Id != "" {
		img.loadImageAsync()
	}
}

func (img *Image) SetShowCaption(show bool) {
	if img.showCaption == show {
		return
	}
	img.showCaption = show
	if img.imageObj != nil {
		img.content = img.imageContent(img.imageObj)
		img.Refresh()
	}
}

// SetAspectRatio is retained for callers using the previous API. Cropping is
// now handled efficiently by canvas.ImageFillCover.
func (img *Image) SetAspectRatio(_, _ int) {}

func (img *Image) createPlaceholderUI() {
	cellBg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	cellBg.StrokeColor = theme.Color(theme.ColorNameSeparator)
	cellBg.StrokeWidth = 1

	icon := widget.NewIcon(theme.MediaPhotoIcon())
	label := widget.NewLabel("Loading thumbnail…")
	label.Alignment = fyne.TextAlignCenter
	label.Importance = widget.LowImportance
	placeholder := container.NewCenter(container.NewVBox(icon, label))

	img.placeholderUI = container.NewStack(cellBg, placeholder)
	img.content = img.placeholderUI
	img.content.Resize(imageTileMinSize)
}

func (img *Image) showPlaceholder() {
	img.content = img.placeholderUI
	img.Refresh()
}

func (img *Image) loadImageAsync() {
	if img.isLoading {
		return
	}
	img.isLoading = true

	picture := img.pic
	go func() {
		loadedImage, err := img.loadImage(picture)

		fyne.Do(func() {
			img.isLoading = false

			if err != nil {
				log.Println("Error loading image:", err)
				if img.pic.Id != picture.Id {
					if img.pic.Id != "" {
						img.loadImageAsync()
					}
					return
				}
				img.showLoadError()
				return
			}
			if img.pic.Id != picture.Id {
				if img.pic.Id != "" {
					img.loadImageAsync()
				}
				return
			}

			canvasImg := canvas.NewImageFromImage(loadedImage)
			canvasImg.FillMode = canvas.ImageFillCover
			canvasImg.SetMinSize(fyne.NewSize(0, 0))
			img.imageObj = canvasImg
			img.content = img.imageContent(canvasImg)
			img.updateBorders()
			img.Refresh()

			if img.onLoad != nil {
				img.onLoad()
			}
		})
	}()
}

func (img *Image) imageContent(canvasImg *canvas.Image) *fyne.Container {
	if !img.showCaption {
		return container.NewStack(canvasImg)
	}
	captionBg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	caption := container.NewStack(captionBg, container.NewPadded(img.caption()))
	return container.NewBorder(nil, caption, nil, nil, canvasImg)
}

func (img *Image) caption() *widget.Label {
	if img.captionLabel == nil {
		img.captionLabel = widget.NewLabel("")
		img.captionLabel.TextStyle = fyne.TextStyle{Bold: true}
		img.captionLabel.Truncation = fyne.TextTruncateEllipsis
	}
	return img.captionLabel
}

func (img *Image) updateCaption(pic datastore.Picture) {
	text := pic.Name
	if text == "" {
		text = filepath.Base(pic.Path)
	}
	if text == "." || text == "" {
		text = "Untitled photo"
	}
	img.caption().SetText(text)
}

func (img *Image) showLoadError() {
	background := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	icon := widget.NewIcon(theme.BrokenImageIcon())
	label := widget.NewLabel("Preview unavailable")
	label.Alignment = fyne.TextAlignCenter
	label.Importance = widget.LowImportance
	img.content = container.NewStack(background, container.NewCenter(container.NewVBox(icon, label)))
	img.Refresh()
}

func (img *Image) loadImage(pic datastore.Picture) (image.Image, error) {
	if img.dataStore == nil {
		return nil, fmt.Errorf("datastore not available")
	}

	const size = "small"
	if file, err := img.dataStore.ImageCache.Get(pic.Id, config.JPEG, size); err == nil {
		loadedImage, _, decodeErr := image.Decode(file)
		_ = file.Close()
		if decodeErr == nil {
			return loadedImage, nil
		}
		log.Printf("Cached thumbnail %s could not be decoded; regenerating it: %v", pic.Id, decodeErr)
	}

	src, err := pic.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load image %s: %w", pic.Id, err)
	}

	var buf bytes.Buffer
	if err := pipeline.ProcessImage(src, config.ImageSizes[size].ImgWidth, config.JPEG, &buf); err != nil {
		return nil, fmt.Errorf("encode processed image %s: %w", pic.Id, err)
	}
	cache, err := img.dataStore.ImageCache.Writer(pic.Id, config.JPEG, size)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache writer: %w", err)
	}
	if _, err := io.Copy(cache, bytes.NewReader(buf.Bytes())); err != nil {
		_ = cache.Abort()
		return nil, fmt.Errorf("cache processed image %s: %w", pic.Id, err)
	}
	if err := cache.Close(); err != nil {
		return nil, fmt.Errorf("publish processed image %s: %w", pic.Id, err)
	}

	loadedImage, _, err := image.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("decode processed image %s: %w", pic.Id, err)
	}
	return loadedImage, nil
}
