package components

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
)

const allPhotosLabel = "All photos"

type ImageGrid struct {
	*datastore.DataStore
	currentPage                                           int
	images, library                                       []datastore.Picture
	itemsPerPage, totalPages                              int
	loading                                               bool
	requestID                                             uint64
	subalbumCount                                         int
	albumID, albumName, selectedID                        string
	grid                                                  *ResponsiveGrid
	tiles                                                 []*Image
	title, statusLabel, pageLabel                         *widget.Label
	previous, next                                        *widget.Button
	search                                                *widget.Entry
	sortOrder                                             *widget.Select
	emptyTitle, emptyHint                                 *widget.Label
	emptyState, loadingState, contentStack, paginationBar *fyne.Container
	layout                                                fyne.CanvasObject
	OnImageSelected                                       func(datastore.Picture)
}

func NewImageGrid(db *datastore.DataStore) *ImageGrid {
	count := config.Config.UI.ImagesPerPage
	if count <= 0 {
		count = 60
	}
	g := &ImageGrid{DataStore: db, itemsPerPage: count, albumName: allPhotosLabel,
		title:       widget.NewLabelWithStyle(allPhotosLabel, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		statusLabel: Hint("Your photo library"), grid: NewResponsiveGrid(180, 1.5, 6, 0)}
	g.title.SizeName = theme.SizeNameHeadingText
	g.title.Truncation = fyne.TextTruncateEllipsis
	g.emptyTitle = widget.NewLabelWithStyle("No photos yet", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	g.emptyHint = Hint("Choose your photo folder in Settings, then scan it from Tasks.")
	g.emptyHint.Alignment = fyne.TextAlignCenter
	g.emptyState = container.NewCenter(container.NewVBox(widget.NewIcon(theme.MediaPhotoIcon()), g.emptyTitle, g.emptyHint))
	g.loadingState = container.NewCenter(widget.NewLabel("Loading photos…"))
	g.contentStack = container.NewStack(g.grid, g.emptyState, g.loadingState)
	g.pageLabel = widget.NewLabel("")
	g.pageLabel.Importance = widget.LowImportance
	g.previous = widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		if g.currentPage > 0 {
			g.currentPage--
			g.Refresh()
		}
	})
	g.next = widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		if g.currentPage+1 < g.totalPages {
			g.currentPage++
			g.Refresh()
		}
	})
	g.paginationBar = container.NewHBox(layout.NewSpacer(), g.previous, g.pageLabel, g.next, layout.NewSpacer())
	g.search = widget.NewEntry()
	g.search.SetPlaceHolder("Search titles, filenames, or tags")
	g.search.OnChanged = func(string) { g.applyFilters() }
	g.sortOrder = widget.NewSelect([]string{"Newest first", "Oldest first", "Title A–Z"}, func(string) { g.applyFilters() })
	g.sortOrder.Selected = "Newest first"
	heading := container.NewVBox(g.title, g.statusLabel)
	toolbar := container.NewBorder(nil, nil, nil, g.sortOrder, g.search)
	g.layout = container.NewBorder(container.NewPadded(container.NewVBox(heading, toolbar)),
		container.NewPadded(g.paginationBar), nil, nil, g.contentStack)
	g.updateState()
	return g
}

// ShowAlbum uses the stable album ID, so duplicate or renamed titles cannot
// merge two folders. Reload preserves the selected album and search.
func (g *ImageGrid) ShowAlbum(id, name string) {
	g.albumID, g.albumName = id, name
	g.title.SetText(name)
	g.selectedID = ""
	g.search.SetText("")
	g.Reload()
}
func (g *ImageGrid) Reload() {
	g.requestID++
	requestID, albumID := g.requestID, g.albumID
	g.SetLoading(true)
	go func() {
		var pictures []datastore.Picture
		var err error
		if albumID == "" {
			pictures, err = g.Pictures.GetAll()
		} else {
			pictures, err = g.Pictures.FindByField("album", albumID)
		}
		fyne.Do(func() {
			if requestID != g.requestID {
				return
			}
			if err != nil {
				g.loading = false
				g.library, g.images = nil, nil
				g.Refresh()
				g.emptyTitle.SetText("Could not load photos")
				g.emptyHint.SetText(err.Error())
				return
			}
			oldPage := g.currentPage
			g.SetImages(pictures)
			g.currentPage = max(0, min(oldPage, g.totalPages-1))
			g.Refresh()
		})
	}()
}
func (g *ImageGrid) SetLoading(loading bool) { g.loading = loading; g.updateState() }
func (g *ImageGrid) SetImages(images []datastore.Picture) {
	g.library = images
	g.loading = false
	g.applyFilters()
}
func (g *ImageGrid) applyFilters() {
	query := strings.ToLower(strings.TrimSpace(g.search.Text))
	g.images = nil
	for _, pic := range g.library {
		text := strings.ToLower(pic.Name + " " + filepath.Base(pic.Path) + " " + pic.Caption + " " + pic.Tags)
		if query == "" || strings.Contains(text, query) {
			g.images = append(g.images, pic)
		}
	}
	sort.SliceStable(g.images, func(i, j int) bool {
		a, b := g.images[i], g.images[j]
		switch g.sortOrder.Selected {
		case "Title A–Z":
			return strings.ToLower(a.Name) < strings.ToLower(b.Name)
		case "Oldest first":
			return a.DateTaken.Before(b.DateTaken)
		default:
			return a.DateTaken.After(b.DateTaken)
		}
	})
	g.currentPage = 0
	g.totalPages = (len(g.images) + g.itemsPerPage - 1) / g.itemsPerPage
	g.Refresh()
}
func (g *ImageGrid) UpdatePicture(pic datastore.Picture) {
	for i := range g.library {
		if g.library[i].Id == pic.Id {
			g.library[i] = pic
		}
	}
	oldPage := g.currentPage
	g.applyFilters()
	g.currentPage = max(0, min(oldPage, g.totalPages-1))
	g.Refresh()
}
func (g *ImageGrid) currentImages() []datastore.Picture {
	start := g.currentPage * g.itemsPerPage
	if start >= len(g.images) {
		return nil
	}
	return g.images[start:min(start+g.itemsPerPage, len(g.images))]
}
func (g *ImageGrid) Layout() fyne.CanvasObject {
	count := config.Config.UI.ImagesPerPage
	if count <= 0 {
		count = 60
	}
	if count != g.itemsPerPage {
		g.itemsPerPage = count
		g.applyFilters()
	}
	return g.layout
}
func (g *ImageGrid) Refresh() {
	g.pageLabel.SetText(fmt.Sprintf("%d / %d", g.currentPage+1, max(1, g.totalPages)))
	g.statusLabel.SetText(PhotoCount(len(g.images)))
	if g.subalbumCount > 0 {
		text := fmt.Sprintf("%d subalbums", g.subalbumCount)
		if g.subalbumCount == 1 {
			text = "1 subalbum"
		}
		if len(g.images) > 0 {
			text += " · " + PhotoCount(len(g.images)) + " in this album"
		}
		g.statusLabel.SetText(text)
	}
	if g.search.Text != "" {
		g.statusLabel.SetText(fmt.Sprintf("%d of %d photos", len(g.images), len(g.library)))
	}
	g.previous.Disable()
	g.next.Disable()
	if g.currentPage > 0 {
		g.previous.Enable()
	}
	if g.currentPage+1 < g.totalPages {
		g.next.Enable()
	}
	pictures := g.currentImages()
	for len(g.tiles) < len(pictures) {
		tile := NewImage(g.DataStore, datastore.Picture{}, nil, func(pic datastore.Picture) {
			g.selectedID = pic.Id
			for _, tile := range g.tiles {
				tile.SetActive(tile.pic.Id == pic.Id)
			}
			if g.OnImageSelected != nil {
				g.OnImageSelected(pic)
			}
		})
		tile.onOpen = func(pic datastore.Picture) { g.OpenPhoto(pic) }
		g.tiles = append(g.tiles, tile)
	}
	objects := make([]fyne.CanvasObject, len(pictures))
	for i, pic := range pictures {
		g.tiles[i].SetPicture(pic)
		g.tiles[i].SetHoverCaption(true)
		g.tiles[i].SetActive(pic.Id == g.selectedID)
		objects[i] = g.tiles[i]
	}
	g.grid.SetObjects(objects)
	g.grid.ScrollToTop()
	g.updateState()
}
func (g *ImageGrid) OpenPhoto(pic datastore.Picture) { ShowPhotoViewer(g.DataStore, g.images, pic.Id) }
func (g *ImageGrid) updateState() {
	defer func() {
		if g.layout != nil {
			g.layout.Refresh()
		}
	}()
	g.loadingState.Hide()
	g.emptyState.Hide()
	g.grid.Hide()
	g.paginationBar.Hide()
	if g.loading {
		g.loadingState.Show()
		return
	}
	if len(g.images) == 0 && g.subalbumCount == 0 {
		if g.search.Text != "" {
			g.emptyTitle.SetText("No matching photos")
			g.emptyHint.SetText("Try a different title, filename, or tag.")
		} else if g.albumID != "" {
			g.emptyTitle.SetText("This album is empty")
			g.emptyHint.SetText("Add photos to this folder and scan again from Tasks.")
		} else {
			g.emptyTitle.SetText("No photos yet")
			g.emptyHint.SetText("Choose your photo folder in Settings, then scan it from Tasks.")
		}
		g.emptyState.Show()
		return
	}
	g.grid.Show()
	if g.totalPages > 1 {
		g.paginationBar.Show()
	}
}

func (g *ImageGrid) SetTitle(title string) { g.albumName = title; g.title.SetText(title) }

// Subalbums share the photo grid's scroller, avoiding competing scroll panes
// when an album contains both folders and photos.
func (g *ImageGrid) SetSubalbums(cards []fyne.CanvasObject) {
	g.subalbumCount = len(cards)
	g.grid.SetLeadingGrid("Subalbums", "Photos", cards, NewResponsiveGridLayoutWithFooter(210, 1.5, 12, 40))
	g.Refresh()
}
