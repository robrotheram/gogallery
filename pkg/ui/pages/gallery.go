package pages

import (
	"gogallery/pkg/datastore"
	"gogallery/pkg/ui/components"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type Page interface {
	Layout() fyne.CanvasObject // Layout returns the main content
}

type GalleryPage struct {
	Title   string
	db      *datastore.DataStore
	sidebar *components.Sidebar
	gallery *components.ImageGrid
	content *fyne.Container
}

func NewGalleryPage(db *datastore.DataStore) *GalleryPage {
	page := &GalleryPage{
		Title: "Gallery",
		db:    db,
	}
	page.sidebar = page.createSideBar()
	// Pass the image selection callback to ImageGrid
	page.gallery = components.NewImageGrid(db)
	page.gallery.OnImageSelected = page.OnImageSelected
	return page
}

// OnImageSelected is called when an image is clicked in the gallery
func (g *GalleryPage) OnImageSelected(img datastore.Picture) {
	cnt := components.NewImageEditContainer(g.db, img)
	g.sidebar.Content = cnt.Layout()
	g.sidebar.Show()
	g.content.Refresh()
}

func (g *GalleryPage) FilterByAlbum(alb string) {
	if pics, err := g.db.Pictures.FindByField("album_name", alb); err == nil {
		g.gallery.SetImages(pics)
	} else {
		log.Println("Error filtering by album:", err)
	}
}

func (g *GalleryPage) createSideBar() *components.Sidebar {
	sidebar := components.NewSidebar("Image Details")
	sidebar.OnToggle = g.refreshLayout

	return sidebar
}

func (g *GalleryPage) refreshLayout() {
	if g.content != nil {
		// Recreate the layout with the current sidebar state
		g.content.Objects = nil
		g.content.Add(container.NewBorder(nil, nil, nil, g.sidebar.Layout(), g.gallery.Layout()))
		g.content.Refresh()
	}
}

func (g *GalleryPage) Refresh() {
	pics, err := g.db.Pictures.GetAll()
	if err != nil {
		panic(err)
	}
	g.gallery.SetImages(pics)
}

func (g *GalleryPage) Layout() fyne.CanvasObject {
	g.Refresh()
	g.content = container.NewBorder(nil, nil, nil, g.sidebar.Layout(), g.gallery.Layout())
	return g.content
}
