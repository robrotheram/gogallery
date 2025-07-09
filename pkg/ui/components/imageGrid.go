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
	"net/http"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ImageGrid struct {
	*datastore.DataStore
	currentPage     int
	images          []datastore.Picture
	itemsPerPage    int
	totalPages      int
	paginationBar   *fyne.Container
	grid            *fyne.Container
	pageLabel       *canvas.Text
	gridItems       []fyne.CanvasObject
	title           *canvas.Text
	selectedAlbum   string                      // Track currently selected album
	OnImageSelected func(pic datastore.Picture) // Callback for image click
}

func NewImageGrid(db *datastore.DataStore) *ImageGrid {
	itemsPerPage := 20 // Default value
	if config.Config.UI.ImagesPerPage > 0 {
		itemsPerPage = config.Config.UI.ImagesPerPage
	}

	ig := &ImageGrid{
		DataStore:    db,
		currentPage:  0,
		itemsPerPage: itemsPerPage,
		totalPages:   0,
		title:        NewTextEntry("All Images", 20),
	}
	ig.pagination()
	ig.imageGrid()

	return ig
}

func (g *ImageGrid) filterByAlbum(alb string) {
	if pics, err := g.DataStore.Pictures.FindByField("album_name", alb); err == nil {
		g.SetImages(pics)
		g.selectedAlbum = alb // Update selected album
		g.title.Text = g.selectedAlbum
	} else {
		log.Println("Error filtering by album:", err)
	}
	g.currentPage = 0 // Reset to first page when filtering
	g.Refresh()
}

func (g *ImageGrid) SetImages(images []datastore.Picture) {
	if len(images) == 0 {
		log.Println("No images to display")
		g.placeholder() // Show placeholder if no images
		g.totalPages = 0
		g.currentPage = 0
		g.Refresh()
		return
	}
	g.totalPages = (len(images) + g.itemsPerPage - 1) / g.itemsPerPage
	g.images = images
	g.currentPage = 0 // Reset to first page when setting new images
	//
	g.Refresh()
}

func (g *ImageGrid) pagination() {
	// Pagination controls
	g.pageLabel = canvas.NewText(fmt.Sprintf("Page %d / %d", g.currentPage+1, g.totalPages), color.White)
	g.pageLabel.Alignment = fyne.TextAlignCenter
	prevBtn := widget.NewButtonWithIcon("Previous", theme.NavigateBackIcon(), func() {
		if g.currentPage > 0 {
			g.currentPage--
			g.Refresh()
		}
	})
	nextBtn := widget.NewButtonWithIcon("Next", theme.NavigateNextIcon(), func() {
		if g.currentPage < g.totalPages-1 {
			g.currentPage++
			g.Refresh()
		}
	})
	nextBtn.IconPlacement = widget.ButtonIconTrailingText
	nextBtn.Alignment = widget.ButtonAlignCenter

	g.paginationBar = container.NewHBox(
		prevBtn,
		layout.NewSpacer(),
		container.NewStack(g.pageLabel),
		layout.NewSpacer(),
		nextBtn,
	)
}

func (g *ImageGrid) imageGrid() {
	// Only create the grid container if it doesn't exist
	if g.grid == nil {
		g.grid = container.New(&ResponsiveGridLayout{
			minCellWidth: 400,
			aspectRatio:  1.5,
			gap:          20,
		}, g.gridItems...)
	} else {
		// Just update the layout, don't reset gridItems
		g.grid.Layout = &ResponsiveGridLayout{
			minCellWidth: 400,
			aspectRatio:  1.5,
			gap:          20,
		}
		g.grid.Refresh()
	}
}

func (g *ImageGrid) placeholder() {
	start := g.currentPage * g.itemsPerPage
	end := min(start+g.itemsPerPage, len(g.images))
	items := make([]fyne.CanvasObject, end-start)
	for i := range items {
		// Use a lightweight placeholder (no URL text, just a rectangle)
		cellBg := canvas.NewRectangle(color.RGBA{R: 241, G: 241, B: 241, A: 255})
		cellBg.StrokeColor = color.Black
		cellBg.StrokeWidth = 1
		// Optionally, add a spinner or "Loading..." label for better UX
		label := canvas.NewText("Loading...", color.Gray{Y: 128})
		label.Alignment = fyne.TextAlignCenter
		cell := container.NewStack(cellBg, label)
		items[i] = cell
	}
	g.gridItems = items
	if g.grid != nil {
		g.grid.Objects = g.gridItems
		g.grid.Refresh()
	}
}

func ImageFromURL(url string) (*canvas.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Failed to fetch image:", err)
		return nil, err
	}
	defer resp.Body.Close()
	imgData, _, err := image.Decode(resp.Body)
	if err != nil {
		log.Println("Failed to decode image:", err)
		return nil, err
	}
	img := canvas.NewImageFromImage(imgData)
	img.FillMode = canvas.ImageFillOriginal
	return img, nil
}

func cropToAspect(imgBuf bytes.Buffer, targetW, targetH int) *bytes.Buffer {
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

func (g *ImageGrid) Thumbnail(pic datastore.Picture) (*canvas.Image, error) {
	size := "small" // Default size

	if file, err := g.ImageCache.Get(pic.Id, config.JPEG, size); err == nil {
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, file); err != nil {
			return nil, fmt.Errorf("failed to read cached image %s: %w", pic.Id, err)
		}
		img := canvas.NewImageFromReader(cropToAspect(buf, 6, 4), "")
		img.FillMode = canvas.ImageFillOriginal // Use Stretch to fill cell, will crop via layout
		img.SetMinSize(fyne.NewSize(0, 0))      // Let layout control size
		return img, nil
	}
	src, err := pic.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load image %s: %w", pic.Id, err)
	}

	cache, err := g.ImageCache.Writer(pic.Id, config.JPEG, size)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache writer: %w", err)
	}

	var buf bytes.Buffer
	multi := io.MultiWriter(cache, &buf)
	pipeline.ProcessImage(src, 400, config.JPEG, multi)

	img := canvas.NewImageFromReader(cropToAspect(buf, 6, 4), "")
	img.FillMode = canvas.ImageFillStretch // Use Stretch to fill cell, will crop via layout
	img.SetMinSize(fyne.NewSize(0, 0))     // Let layout control size
	return img, nil
}

// Load replace the placeholder with actual images
// This function can be used to load images asynchronously or on demand.
func (g *ImageGrid) LoadImages() {
	g.placeholder()

	start := g.currentPage * g.itemsPerPage
	end := min((g.currentPage+1)*g.itemsPerPage, len(g.images))
	for i := start; i < end; i++ {
		idx := i - start
		pic := g.images[i]
		go func(i, idx int, pic datastore.Picture) {
			img, err := g.Thumbnail(pic)
			if err != nil {
				log.Println("Error loading image:", err)
				return
			}
			// Make the image fill the cell and be fully clickable, with a border on hover
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(180, 180)) // Adjust as needed for your grid cell size
			cell := NewImageCell(img, func() {
				if g.OnImageSelected != nil {
					g.OnImageSelected(pic)
				}
			})
			if idx < len(g.grid.Objects) {
				g.grid.Objects[idx] = cell
			}
			fyne.Do(func() {
				g.grid.Refresh()
			})
		}(i, idx, pic)
	}
	log.Println("Started loading images for page", g.currentPage+1)
}
func (g *ImageGrid) galleryHeader() fyne.CanvasObject {
	leftPad := canvas.NewRectangle(nil)
	leftPad.SetMinSize(fyne.NewSize(12, 0))
	const allPhotosLabel = "All Photos"

	albms, _ := g.Albums.GetLatestAlbums()
	albumOptions := make([]string, len(albms)+1)
	albumOptions[0] = allPhotosLabel // First option for all photos
	for i, album := range albms {
		albumOptions[i+1] = album.Name
	}

	albumSelect := widget.NewSelect(albumOptions, func(selected string) {
		log.Printf("Selected album: %s", selected)
		if selected == allPhotosLabel {
			pics, err := g.DataStore.Pictures.GetAll()
			if err != nil {
				return
			}
			g.SetImages(pics)
			return
		}
		g.filterByAlbum(selected)
	})
	albumSelect.PlaceHolder = allPhotosLabel
	return (container.NewHBox(leftPad, albumSelect))
}

func (g *ImageGrid) Layout() fyne.CanvasObject {
	stack := container.NewVBox(g.galleryHeader(), g.grid)
	return container.NewBorder(
		nil,                         // top
		g.paginationBar,             // bottom (footer)
		nil,                         // left
		nil,                         // right
		container.NewVScroll(stack), // center (main scroll area)
	)
}

func (g *ImageGrid) Refresh() {
	g.pageLabel.Text = fmt.Sprintf("Page %d / %d", g.currentPage+1, g.totalPages)
	g.pageLabel.Refresh()
	g.LoadImages()
}
