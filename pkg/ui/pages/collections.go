package pages

import (
	"gogallery/pkg/datastore"
	"gogallery/pkg/ui/components"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type CollectionPage struct {
	*datastore.DataStore
	Title   string
	albums  []datastore.Album
	grid    *fyne.Container
	sidebar *components.Sidebar
	content *fyne.Container
}

type albumCardData struct {
	album   datastore.Album
	picture datastore.Picture
}

func NewCollectionPage(db *datastore.DataStore) *CollectionPage {
	page := &CollectionPage{
		DataStore: db,
		Title:     "Collections",
		albums:    []datastore.Album{},
		grid:      container.New(components.NewResponsiveGridLayout(500, 1.3, 20)),
	}
	page.sidebar = components.NewSidebar("Album Settings")
	page.sidebar.OnToggle = page.RefreshLayout
	page.content = container.NewBorder(nil, nil, nil, page.sidebar.Layout(), container.NewVScroll(page.grid))
	return page
}

func (page *CollectionPage) RefreshLayout() {
	if page.content != nil {
		// Recreate the layout with the current sidebar state
		page.content.Objects = nil
		page.content.Add(container.NewBorder(nil, nil, nil, page.sidebar.Layout(), container.NewVScroll(page.grid)))
		page.content.Refresh()
	}
}

func (page *CollectionPage) Refresh() {
	log.Println("Refreshing CollectionPage")

	// Do database operations in background
	go func() {
		albums, err := page.Albums.GetAll()
		if err != nil {
			log.Printf("Error loading albums: %v", err)
			return
		}

		// Filter out blacklisted albums
		var filteredAlbums []datastore.Album
		for _, alb := range albums {
			if !datastore.IsAlbumInBlacklist(alb.Name) {
				filteredAlbums = append(filteredAlbums, alb)
			}
		}

		cardData := make([]albumCardData, 0, len(filteredAlbums))
		for _, alb := range filteredAlbums {
			if alb.ProfileId == "" {
				continue
			}
			picture, err := page.Pictures.FindByID(alb.ProfileId)
			if err != nil {
				continue
			}
			cardData = append(cardData, albumCardData{album: alb, picture: picture})
		}

		// Update UI on main thread
		fyne.Do(func() {
			cells := make([]fyne.CanvasObject, 0, len(cardData))
			for _, data := range cardData {
				cells = append(cells, page.makeAlbumCard(data.album, data.picture))
			}
			page.grid.Objects = cells
			page.grid.Refresh()
		})
	}()
}

func (page *CollectionPage) Layout() fyne.CanvasObject {
	// Refresh asynchronously to avoid blocking UI
	page.Refresh()
	return page.content
}

func (page *CollectionPage) makeAlbumCard(alb datastore.Album, pic datastore.Picture) fyne.CanvasObject {
	img := components.NewImage(page.DataStore, pic, nil, func(clickedPic datastore.Picture) {
		log.Printf("Album %s clicked", alb.Name)
		cnt := components.NewCollectionAlbumContainer(page.DataStore, alb)
		cnt.OnUpdate = func() {
			page.Refresh()
		}
		page.sidebar.Content = cnt.Layout()
		page.sidebar.Show()
	})

	// Create the album name label for the footer
	nameLabel := widget.NewLabel(alb.Name)
	nameLabel.TextStyle = fyne.TextStyle{Bold: true}
	nameLabel.Alignment = fyne.TextAlignCenter

	// Create the card with a border layout - image in center, name label as footer
	card := container.NewBorder(
		nil,       // top
		nameLabel, // bottom (footer with album name)
		nil,       // left
		nil,       // right
		img,       // center (main image)
	)
	return card
}
