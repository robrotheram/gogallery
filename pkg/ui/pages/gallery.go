package pages

import (
	"fyne.io/fyne/v2"
	"gogallery/pkg/datastore"
	"gogallery/pkg/ui/components"
)

type Page interface{ Layout() fyne.CanvasObject }
type GalleryPage struct {
	Title          string
	db             *datastore.DataStore
	sidebar        *components.Sidebar
	gallery        *components.ImageGrid
	content        *fyne.Container
	loaded         bool
	OnAlbumUpdated func()
	OnPhotoUpdated func()
}

func NewGalleryPage(db *datastore.DataStore) *GalleryPage {
	page := &GalleryPage{Title: "Photos", db: db, sidebar: components.NewSidebar("Photo details"), gallery: components.NewImageGrid(db)}
	page.gallery.OnImageSelected = page.OnImageSelected
	page.content = components.NewInspectorWorkspace(page.gallery.Layout(), page.sidebar)
	return page
}
func (g *GalleryPage) OnImageSelected(pic datastore.Picture) {
	editor := components.NewImageEditContainer(g.db, pic)
	editor.OnUpdate = func(updated datastore.Picture) {
		g.gallery.UpdatePicture(updated)
		if g.OnPhotoUpdated != nil {
			g.OnPhotoUpdated()
		}
	}
	editor.OnOpen = func() { g.gallery.OpenPhoto(pic) }
	editor.OnCoverChanged = func() {
		if g.OnAlbumUpdated != nil {
			g.OnAlbumUpdated()
		}
	}
	g.sidebar.SetContent("Photo details", editor.Layout())
}
func (g *GalleryPage) Refresh() { g.loaded = true; g.gallery.Reload() }
func (g *GalleryPage) Layout() fyne.CanvasObject {
	g.gallery.Layout()
	if !g.loaded {
		g.Refresh()
	}
	return g.content
}
