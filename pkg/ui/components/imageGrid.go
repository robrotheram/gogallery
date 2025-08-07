package components

import (
	"fmt"
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"image"
	"image/color"
	"log"
	"net/http"
	"sort"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ImageGrid struct {
	*datastore.DataStore
	currentPage   int
	images        []datastore.Picture
	itemsPerPage  int
	totalPages    int
	paginationBar *fyne.Container
	grid          *fyne.Container
	pageLabel     *canvas.Text
	// gridItems       []fyne.CanvasObject
	title           *canvas.Text
	selectedAlbum   string                      // Track currently selected album
	OnImageSelected func(pic datastore.Picture) // Callback for image click

	// Image caching for better navigation performance
	imageCache     map[string]*Image // Cache of loaded Image widgets by picture ID
	cacheMutex     sync.RWMutex      // Mutex to protect the cache
	preloadWorkers int               // Number of background workers for preloading
	refreshMutex   sync.Mutex        // Mutex to prevent concurrent refreshes
}

func NewImageGrid(db *datastore.DataStore) *ImageGrid {
	itemsPerPage := 20 // Default value
	if config.Config.UI.ImagesPerPage > 0 {
		itemsPerPage = config.Config.UI.ImagesPerPage
	}

	ig := &ImageGrid{
		DataStore:      db,
		currentPage:    0,
		itemsPerPage:   itemsPerPage,
		totalPages:     0,
		title:          NewTextEntry("All Images", 20),
		imageCache:     make(map[string]*Image),
		preloadWorkers: 2, // Number of concurrent preloading workers
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
		// g.placeholder() // Show placeholder if no images
		g.totalPages = 0
		g.currentPage = 0
		// Make refresh async to avoid blocking
		go g.Refresh()
		return
	}
	g.totalPages = (len(images) + g.itemsPerPage - 1) / g.itemsPerPage
	g.images = images
	g.currentPage = 0 // Reset to first page when setting new images

	// Clear the image cache when new images are set
	g.clearImageCache()

	// Make refresh async to avoid blocking
	go g.Refresh()
}

func (g *ImageGrid) pagination() {
	// Pagination controls
	g.pageLabel = canvas.NewText(fmt.Sprintf("Page %d / %d", g.currentPage+1, g.totalPages), color.White)
	g.pageLabel.Alignment = fyne.TextAlignCenter
	prevBtn := widget.NewButtonWithIcon("Previous", theme.NavigateBackIcon(), func() {
		if g.refreshMutex.TryLock() {
			defer g.refreshMutex.Unlock()
			if g.currentPage > 0 {
				g.currentPage--
				go g.Refresh() // Make refresh async to avoid blocking UI
			}
		}
	})
	nextBtn := widget.NewButtonWithIcon("Next", theme.NavigateNextIcon(), func() {
		if g.refreshMutex.TryLock() {
			defer g.refreshMutex.Unlock()
			if g.currentPage < g.totalPages-1 {
				g.currentPage++
				go g.Refresh() // Make refresh async to avoid blocking UI
			}
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
		g.grid = container.New(NewResponsiveGridLayout(400, 1.5, 20))
	} else {
		// Just update the layout, don't reset gridItems
		g.grid.Layout = NewResponsiveGridLayout(400, 1.5, 20)
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

func (g *ImageGrid) LoadImages() {
	start := g.currentPage * g.itemsPerPage
	end := min((g.currentPage+1)*g.itemsPerPage, len(g.images))

	// Create cells for current page, using cache when available
	cells := make([]fyne.CanvasObject, end-start)
	for i := range cells {
		idx := start + i // Correct index from the current page
		pic := g.images[idx]

		// Try to get from cache first
		if cachedImg := g.getCachedImage(pic.Id); cachedImg != nil {
			cells[i] = cachedImg
		} else {
			// Create new image and cache it
			newImg := NewImage(g.DataStore, pic, func() {
				// Use fyne.Do to ensure grid refresh happens on UI thread
				fyne.Do(func() {
					g.grid.Refresh()
				})
			}, func(clickedPic datastore.Picture) {
				if g.OnImageSelected != nil {
					g.OnImageSelected(clickedPic)
				}
			})
			g.cacheImage(pic.Id, newImg)
			cells[i] = newImg
		}
	}

	// Update UI on main thread
	fyne.Do(func() {
		g.grid.Objects = cells
		g.grid.Refresh()
	})

	log.Println("Started loading images for page", g.currentPage+1)

	// Start preloading adjacent pages in background
	go g.preloadAdjacentPages()
}

type TappableContainer struct {
	*fyne.Container
	onTap func()
}

func (t *TappableContainer) Tapped(_ *fyne.PointEvent) {
	if t.onTap != nil {
		t.onTap()
	}
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
	sort.Strings(albumOptions[1:]) // Sort album options alphabetically, excluding "All Photos"

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

// clearImageCache clears all cached images
func (g *ImageGrid) clearImageCache() {
	g.cacheMutex.Lock()
	defer g.cacheMutex.Unlock()
	g.imageCache = make(map[string]*Image)
	log.Println("Image cache cleared")
}

// getCachedImage retrieves an image from cache
func (g *ImageGrid) getCachedImage(picId string) *Image {
	g.cacheMutex.RLock()
	defer g.cacheMutex.RUnlock()
	return g.imageCache[picId]
}

// cacheImage stores an image in cache
func (g *ImageGrid) cacheImage(picId string, img *Image) {
	g.cacheMutex.Lock()
	defer g.cacheMutex.Unlock()
	g.imageCache[picId] = img
}

// preloadAdjacentPages preloads images from previous and next pages
func (g *ImageGrid) preloadAdjacentPages() {
	if len(g.images) == 0 {
		return
	}

	// Preload previous page if it exists
	if g.currentPage > 0 {
		go g.preloadPage(g.currentPage - 1)
	}

	// Preload next page if it exists
	if g.currentPage < g.totalPages-1 {
		go g.preloadPage(g.currentPage + 1)
	}
}

// preloadPage preloads images for a specific page
func (g *ImageGrid) preloadPage(pageNum int) {
	if pageNum < 0 || pageNum >= g.totalPages {
		return
	}

	start := pageNum * g.itemsPerPage
	end := min((pageNum+1)*g.itemsPerPage, len(g.images))

	// Use a semaphore to limit concurrent preloading
	semaphore := make(chan struct{}, g.preloadWorkers)

	for i := start; i < end; i++ {
		pic := g.images[i]

		// Skip if already cached
		if g.getCachedImage(pic.Id) != nil {
			continue
		}

		semaphore <- struct{}{} // Acquire
		go func(picture datastore.Picture) {
			defer func() { <-semaphore }() // Release

			// Create and cache the image
			img := NewImage(g.DataStore, picture, func() {
				// No need to refresh grid for preloaded images
			}, func(clickedPic datastore.Picture) {
				if g.OnImageSelected != nil {
					g.OnImageSelected(clickedPic)
				}
			})
			g.cacheImage(picture.Id, img)
		}(pic)
	}
}
