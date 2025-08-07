package components

import (
	"bytes"
	"fmt"
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"gogallery/pkg/pipeline"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// Image is a custom Fyne widget for displaying images with placeholder and async loading
type Image struct {
	widget.BaseWidget

	dataStore *datastore.DataStore
	pic       datastore.Picture

	// UI components
	content       *fyne.Container
	placeholderUI *fyne.Container
	imageObj      *canvas.Image
	hoverBorder   *canvas.Rectangle
	activeBorder  *canvas.Rectangle
	isLoading     bool
	isHovered     bool
	isActive      bool
	onLoad        func()                  // Callback for when image is loaded
	onClick       func(datastore.Picture) // Callback for when image is clicked

	// Aspect ratio for cropping (width, height)
	aspectWidth  int
	aspectHeight int
}

func NewImage(dataStore *datastore.DataStore, pic datastore.Picture, onLoad func(), onClick func(datastore.Picture)) *Image {
	return NewImageWithAspect(dataStore, pic, onLoad, onClick, 6, 4) // Default to 6:4 aspect ratio
}

func NewImageWithAspect(dataStore *datastore.DataStore, pic datastore.Picture, onLoad func(), onClick func(datastore.Picture), aspectWidth, aspectHeight int) *Image {
	img := &Image{
		dataStore:    dataStore,
		pic:          pic,
		isLoading:    false,
		isHovered:    false,
		isActive:     false,
		onLoad:       onLoad,
		onClick:      onClick,
		aspectWidth:  aspectWidth,
		aspectHeight: aspectHeight,
	}
	img.ExtendBaseWidget(img)
	img.createBorders()
	img.createPlaceholderUI()

	// Start loading the image asynchronously if we have a valid picture
	if pic.Id != "" {
		go img.loadImageAsync()
	}

	return img
}

// Tapped implements the fyne.Tappable interface
func (img *Image) Tapped(_ *fyne.PointEvent) {
	if img.onClick != nil {
		img.onClick(img.pic)
	}
}

// MouseIn implements the desktop.Hoverable interface
func (img *Image) MouseIn(_ *desktop.MouseEvent) {
	img.isHovered = true
	img.updateBorders()
}

// MouseOut implements the desktop.Hoverable interface
func (img *Image) MouseOut() {
	img.isHovered = false
	img.updateBorders()
}

// MouseMoved implements the desktop.Hoverable interface
func (img *Image) MouseMoved(_ *desktop.MouseEvent) {
	// No action needed for mouse movement
}

// Cursor returns the cursor to show when hovering
func (img *Image) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

func (img *Image) createBorders() {
	// Hover border - blue with transparency
	img.hoverBorder = canvas.NewRectangle(color.NRGBA{R: 0, G: 120, B: 255, A: 100})
	img.hoverBorder.StrokeWidth = 3
	img.hoverBorder.StrokeColor = color.NRGBA{R: 0, G: 120, B: 255, A: 255}
	img.hoverBorder.Hide()

	// Active border - darker blue
	img.activeBorder = canvas.NewRectangle(color.NRGBA{R: 0, G: 80, B: 200, A: 150})
	img.activeBorder.StrokeWidth = 4
	img.activeBorder.StrokeColor = color.NRGBA{R: 0, G: 80, B: 200, A: 255}
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
	img.Refresh()
}

func (img *Image) CreateRenderer() fyne.WidgetRenderer {
	if img.content == nil {
		img.content = img.placeholderUI
	}
	return &imageRenderer{
		image: img,
	}
}

// Custom renderer for the Image widget
type imageRenderer struct {
	image   *Image
	objects []fyne.CanvasObject
}

func (r *imageRenderer) Layout(size fyne.Size) {
	if r.image.content != nil {
		r.image.content.Resize(size)
	}

	// Position borders to cover the entire widget area
	if r.image.hoverBorder != nil {
		r.image.hoverBorder.Resize(size)
	}
	if r.image.activeBorder != nil {
		r.image.activeBorder.Resize(size)
	}
}

func (r *imageRenderer) MinSize() fyne.Size {
	if r.image.content != nil {
		return r.image.content.MinSize()
	}
	return fyne.NewSize(100, 100) // Default minimum size
}

func (r *imageRenderer) Refresh() {
	// Update the objects list to reflect current content
	if r.image.content != nil {
		r.objects = []fyne.CanvasObject{r.image.content}
		r.image.content.Refresh()
	} else {
		r.objects = []fyne.CanvasObject{}
	}
}

func (r *imageRenderer) Objects() []fyne.CanvasObject {
	// Always return the current content plus borders
	objects := []fyne.CanvasObject{}

	if r.image.content != nil {
		objects = append(objects, r.image.content)
	}

	// Add borders
	if r.image.hoverBorder != nil {
		objects = append(objects, r.image.hoverBorder)
	}
	if r.image.activeBorder != nil {
		objects = append(objects, r.image.activeBorder)
	}

	return objects
}

func (r *imageRenderer) Destroy() {
	// Cleanup if needed
}

func (img *Image) SetPicture(pic datastore.Picture) {
	img.pic = pic
	img.showPlaceholder()

	if pic.Id == "" {
		return // No picture set, just show placeholder
	}

	// Load image asynchronously
	go img.loadImageAsync()
}

func (img *Image) SetAspectRatio(width, height int) {
	img.aspectWidth = width
	img.aspectHeight = height

	// If we have a loaded image, reload it with the new aspect ratio
	if img.pic.Id != "" && !img.isLoading {
		go img.loadImageAsync()
	}
}

func (img *Image) createPlaceholderUI() {

	cellBg := canvas.NewRectangle(color.RGBA{R: 241, G: 241, B: 241, A: 255})
	cellBg.StrokeColor = color.Black
	cellBg.StrokeWidth = 1

	label := canvas.NewText("Loading...", color.Gray{Y: 128})
	label.Alignment = fyne.TextAlignCenter

	// Create container with content and borders
	img.placeholderUI = container.NewStack(cellBg, label)
	img.content = img.placeholderUI
	img.content.Resize(fyne.NewSize(200, 150)) // Default size for placeholder
}

func (img *Image) showPlaceholder() {
	img.content = img.placeholderUI
	img.Refresh()
}

func (img *Image) loadImageAsync() {
	if img.isLoading {
		log.Println("Image is already loading, skipping")
		return // Already loading
	}
	img.isLoading = true

	// Load and process image in background thread
	go func() {
		canvasImg, err := img.loadImage(img.pic)

		// Update UI on main thread
		fyne.Do(func() {
			img.isLoading = false

			if err != nil {
				// Keep showing placeholder on error
				log.Println("Error loading image:", err)
				return
			}

			// Update UI components
			img.imageObj = canvasImg
			img.content = container.NewStack(img.imageObj)

			// Update borders visibility based on current state
			img.updateBorders()

			img.Refresh() // This will now properly trigger the custom renderer

			if img.onLoad != nil {
				img.onLoad() // Call the onLoad callback if set
			}
		})
	}()
}

func (img *Image) cropToAspect(imgBuf bytes.Buffer, targetW, targetH int) *bytes.Buffer {
	// Decode image from buffer
	srcImg, _, err := image.Decode(&imgBuf)
	if err != nil {
		return &imgBuf // fallback: return original if decode fails
	}
	srcBounds := srcImg.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	targetAspect := float64(targetW) / float64(targetH)
	srcAspect := float64(srcW) / float64(srcH)

	var cropW, cropH int
	if srcAspect > targetAspect {
		// Source is wider than target: crop width
		cropH = srcH
		cropW = int(float64(cropH) * targetAspect)
	} else {
		// Source is taller than target: crop height
		cropW = srcW
		cropH = int(float64(cropW) / targetAspect)
	}
	x0 := srcBounds.Min.X + (srcW-cropW)/2
	y0 := srcBounds.Min.Y + (srcH-cropH)/2
	cropRect := image.Rect(x0, y0, x0+cropW, y0+cropH)

	// Crop and copy to a new RGBA image
	cropped := image.NewRGBA(image.Rect(0, 0, cropW, cropH))
	draw.Draw(cropped, cropped.Bounds(), srcImg, cropRect.Min, draw.Src)

	// Encode cropped image back to buffer
	var outBuf bytes.Buffer
	jpeg.Encode(&outBuf, cropped, nil)
	return &outBuf
}

func (img *Image) loadImage(pic datastore.Picture) (*canvas.Image, error) {
	if img.dataStore == nil {
		return nil, fmt.Errorf("datastore not available")
	}

	size := "small" // Default size

	// Try to get from cache first
	if file, err := img.dataStore.ImageCache.Get(pic.Id, config.JPEG, size); err == nil {
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, file); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to read cached image %s: %w", pic.Id, err)
		}
		file.Close()

		// Do cropping in background thread
		croppedBuf := img.cropToAspect(buf, img.aspectWidth, img.aspectHeight)
		canvasImg := canvas.NewImageFromReader(croppedBuf, "")
		canvasImg.FillMode = canvas.ImageFill(canvas.ImageScaleFastest)
		canvasImg.SetMinSize(fyne.NewSize(0, 0))
		return canvasImg, nil
	}

	log.Println("Image not in cache, loading from source:", pic.Id)
	// Load from source and cache
	src, err := pic.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load image %s: %w", pic.Id, err)
	}

	cache, err := img.dataStore.ImageCache.Writer(pic.Id, config.JPEG, size)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache writer: %w", err)
	}
	defer cache.Close()

	var buf bytes.Buffer
	multi := io.MultiWriter(cache, &buf)
	pipeline.ProcessImage(src, 400, config.JPEG, multi)

	// Do cropping in background thread
	croppedBuf := img.cropToAspect(buf, img.aspectWidth, img.aspectHeight)
	canvasImg := canvas.NewImageFromReader(croppedBuf, "")
	canvasImg.FillMode = canvas.ImageFillStretch
	canvasImg.SetMinSize(fyne.NewSize(0, 0))
	log.Println("Successfully loaded and cached image:", pic.Id)
	return canvasImg, nil
}
