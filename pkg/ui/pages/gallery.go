package pages

import (
	"log"
	"testingFyne/pkg/datastore"
	"testingFyne/pkg/ui/components"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type Page interface {
	Layout() fyne.CanvasObject // Layout returns the main content
}

type GalleryPage struct {
	Title         string
	db            *datastore.DataStore
	sidebar       *components.Sidebar
	gallery       *components.ImageGrid
	content       *fyne.Container
	galleryWidget fyne.CanvasObject
	sidebarWidget fyne.CanvasObject
}

func NewGalleryPage(db *datastore.DataStore) *GalleryPage {
	page := &GalleryPage{
		Title: "Gallery",
		db:    db,
	}
	page.sidebar = components.NewSidebar(db, page.CloseSidebar)
	// Pass the image selection callback to ImageGrid
	page.gallery = components.NewImageGrid(db)
	page.gallery.OnImageSelected = page.OnImageSelected
	return page
}

// OnImageSelected is called when an image is clicked in the gallery
func (g *GalleryPage) OnImageSelected(img datastore.Picture) {
	g.sidebar.ShowImage(img)
	if g.content != nil {
		sidebarBox := container.NewStack(g.sidebarWidget)
		g.content.Objects = []fyne.CanvasObject{
			container.NewBorder(nil, nil, nil, sidebarBox, g.galleryWidget),
		}
		g.content.Refresh()
	}
}

func (g *GalleryPage) FilterByAlbum(alb string) {
	if pics, err := g.db.Pictures.FindByField("album_name", alb); err == nil {
		g.gallery.SetImages(pics)
	} else {
		log.Println("Error filtering by album:", err)
	}
}

func (g *GalleryPage) CloseSidebar() {
	g.sidebar.Hide()
	if g.content != nil {
		g.content.Objects = []fyne.CanvasObject{
			container.NewBorder(nil, nil, nil, nil, g.galleryWidget),
		}
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
	g.galleryWidget = g.gallery.Layout()
	g.sidebarWidget = g.sidebar.Layout()
	g.content = container.NewBorder(nil, nil, nil, nil, g.galleryWidget)
	return g.content
}
