package components

import (
	"fmt"
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"log"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const allPhotosLabel = "All Photos"

type ImageGrid struct {
	*datastore.DataStore
	currentPage  int
	images       []datastore.Picture
	itemsPerPage int
	totalPages   int
	loading      bool
	requestID    uint64

	grid          *ResponsiveGrid
	tiles         []*Image
	title         *widget.Label
	statusLabel   *widget.Label
	pageLabel     *widget.Label
	previous      *widget.Button
	next          *widget.Button
	emptyState    *fyne.Container
	loadingState  *fyne.Container
	contentStack  *fyne.Container
	paginationBar *fyne.Container
	layout        fyne.CanvasObject

	OnImageSelected func(pic datastore.Picture)
}

func NewImageGrid(db *datastore.DataStore) *ImageGrid {
	itemsPerPage := 20
	if config.Config.UI.ImagesPerPage > 0 {
		itemsPerPage = config.Config.UI.ImagesPerPage
	}

	g := &ImageGrid{
		DataStore:    db,
		itemsPerPage: itemsPerPage,
		title:        widget.NewLabelWithStyle(allPhotosLabel, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		statusLabel:  widget.NewLabel("Loading library…"),
	}
	g.statusLabel.Importance = widget.MediumImportance

	// Page size bounds the number of tile widgets while the custom layout lets
	// every row expand to the exact viewport width.
	g.grid = NewResponsiveGrid(230, 1.5, 8, 48)

	g.createStates()
	g.pagination()
	g.layout = g.buildLayout()
	g.updateState()
	return g
}

func (g *ImageGrid) createStates() {
	emptyIcon := widget.NewIcon(theme.MediaPhotoIcon())
	emptyTitle := widget.NewLabelWithStyle("No photos found", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	emptyHint := widget.NewLabel("Choose another album or add photos to your gallery folder.")
	emptyHint.Alignment = fyne.TextAlignCenter
	emptyHint.Importance = widget.LowImportance
	g.emptyState = container.NewCenter(container.NewVBox(emptyIcon, emptyTitle, emptyHint))

	progress := widget.NewProgressBarInfinite()
	loadingLabel := widget.NewLabel("Loading photos…")
	loadingLabel.Alignment = fyne.TextAlignCenter
	loadingLabel.Importance = widget.LowImportance
	g.loadingState = container.NewCenter(container.NewVBox(progress, loadingLabel))

	g.contentStack = container.NewStack(g.grid, g.emptyState, g.loadingState)
}

func (g *ImageGrid) pagination() {
	g.pageLabel = widget.NewLabel("")
	g.pageLabel.Alignment = fyne.TextAlignCenter
	g.pageLabel.Importance = widget.LowImportance

	g.previous = widget.NewButtonWithIcon("Previous", theme.NavigateBackIcon(), func() {
		if g.currentPage > 0 {
			g.currentPage--
			g.Refresh()
		}
	})
	g.next = widget.NewButtonWithIcon("Next", theme.NavigateNextIcon(), func() {
		if g.currentPage < g.totalPages-1 {
			g.currentPage++
			g.Refresh()
		}
	})
	g.next.IconPlacement = widget.ButtonIconTrailingText

	g.paginationBar = container.NewHBox(
		g.previous,
		layout.NewSpacer(),
		g.pageLabel,
		layout.NewSpacer(),
		g.next,
	)
}

func (g *ImageGrid) buildLayout() fyne.CanvasObject {
	albumSelect := widget.NewSelect(g.albumOptions(), func(selected string) {
		if selected == "" {
			return
		}
		log.Printf("Selected album: %s", selected)
		g.loadAlbum(selected)
	})
	albumSelect.Selected = allPhotosLabel
	albumSelect.PlaceHolder = allPhotosLabel

	titleBlock := container.NewHBox(g.title, g.statusLabel)
	filter := container.NewHBox(widget.NewLabel("Album"), albumSelect)
	header := container.NewBorder(nil, nil, titleBlock, filter)

	body := container.NewBorder(
		container.NewPadded(header),
		container.NewPadded(g.paginationBar),
		nil,
		nil,
		container.NewPadded(g.contentStack),
	)
	return body
}

func (g *ImageGrid) albumOptions() []string {
	options := []string{allPhotosLabel}
	albums, err := g.Albums.GetLatestAlbums()
	if err != nil {
		log.Printf("Error loading albums: %v", err)
		return options
	}
	for _, album := range albums {
		options = append(options, album.Name)
	}
	sort.Strings(options[1:])
	return options
}

func (g *ImageGrid) loadAlbum(album string) {
	g.requestID++
	requestID := g.requestID
	g.SetLoading(true)

	go func() {
		var (
			pictures []datastore.Picture
			err      error
		)
		if album == allPhotosLabel {
			pictures, err = g.DataStore.Pictures.GetAll()
		} else {
			pictures, err = g.DataStore.Pictures.FindByField("album_name", album)
		}

		fyne.Do(func() {
			if requestID != g.requestID {
				return
			}
			if err != nil {
				log.Printf("Error loading album %q: %v", album, err)
				g.SetLoading(false)
				return
			}
			g.title.SetText(album)
			g.SetImages(pictures)
		})
	}()
}

// Reload refreshes the full library while preserving the grid's async loading
// and request ordering guarantees.
func (g *ImageGrid) Reload() {
	g.loadAlbum(allPhotosLabel)
}

func (g *ImageGrid) SetLoading(loading bool) {
	g.loading = loading
	g.updateState()
}

func (g *ImageGrid) SetImages(images []datastore.Picture) {
	g.images = images
	g.currentPage = 0
	if len(images) == 0 {
		g.totalPages = 0
	} else {
		g.totalPages = (len(images) + g.itemsPerPage - 1) / g.itemsPerPage
	}
	g.loading = false
	g.Refresh()
}

func (g *ImageGrid) currentImages() []datastore.Picture {
	start := g.currentPage * g.itemsPerPage
	if start >= len(g.images) {
		return nil
	}
	end := min(start+g.itemsPerPage, len(g.images))
	return g.images[start:end]
}

func (g *ImageGrid) Layout() fyne.CanvasObject {
	return g.layout
}

func (g *ImageGrid) Refresh() {
	pageNumber := 0
	if g.totalPages > 0 {
		pageNumber = g.currentPage + 1
	}
	g.pageLabel.SetText(fmt.Sprintf("Page %d of %d", pageNumber, g.totalPages))

	start := 0
	end := 0
	if len(g.images) > 0 {
		start = g.currentPage*g.itemsPerPage + 1
		end = min((g.currentPage+1)*g.itemsPerPage, len(g.images))
	}
	g.statusLabel.SetText(fmt.Sprintf("Showing %d–%d of %d photos", start, end, len(g.images)))

	if g.currentPage == 0 {
		g.previous.Disable()
	} else {
		g.previous.Enable()
	}
	if g.totalPages == 0 || g.currentPage >= g.totalPages-1 {
		g.next.Disable()
	} else {
		g.next.Enable()
	}

	g.updateTiles()
	g.grid.ScrollToTop()
	g.updateState()
}

func (g *ImageGrid) updateTiles() {
	pictures := g.currentImages()
	for len(g.tiles) < len(pictures) {
		image := NewImage(g.DataStore, datastore.Picture{}, nil, func(pic datastore.Picture) {
			if g.OnImageSelected != nil {
				g.OnImageSelected(pic)
			}
		})
		image.SetShowCaption(true)
		g.tiles = append(g.tiles, image)
	}

	objects := make([]fyne.CanvasObject, len(pictures))
	for i, picture := range pictures {
		g.tiles[i].SetPicture(picture)
		objects[i] = g.tiles[i]
	}
	g.grid.SetObjects(objects)
}

func (g *ImageGrid) updateState() {
	if g.loading {
		g.grid.Hide()
		g.emptyState.Hide()
		g.loadingState.Show()
		g.paginationBar.Hide()
		return
	}

	g.loadingState.Hide()
	if len(g.images) == 0 {
		g.grid.Hide()
		g.emptyState.Show()
		g.paginationBar.Hide()
		return
	}

	g.emptyState.Hide()
	g.grid.Show()
	g.paginationBar.Show()
}
