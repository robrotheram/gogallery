package pages

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gogallery/pkg/datastore"
	"gogallery/pkg/ui/components"
	"strings"
)

type CollectionPage struct {
	*datastore.DataStore
	Title            string
	albums           []albumCardData
	grid             *components.ResponsiveGrid
	sidebar          *components.Sidebar
	photos           *components.ImageGrid
	content, view    *fyne.Container
	breadcrumbs      *fyne.Container
	back             *widget.Button
	overview, detail fyne.CanvasObject
	search           *widget.Entry
	status, empty    *widget.Label
	selected         *datastore.Album
	requestID        uint64
	loaded           bool
	OnPhotoUpdated   func()
}

func NewCollectionPage(db *datastore.DataStore) *CollectionPage {
	p := &CollectionPage{DataStore: db, Title: "Albums", grid: components.NewResponsiveGrid(210, 1.5, 12, 40), sidebar: components.NewSidebar("Album details")}
	p.status = components.Hint("Albums follow the folders in your photo library.")
	p.search = widget.NewEntry()
	p.search.SetPlaceHolder("Find an album")
	p.search.OnChanged = func(string) { p.filterAlbums() }
	p.empty = components.Hint("Loading albums…")
	p.empty.Alignment = fyne.TextAlignCenter
	title := components.NewTextEntry("Albums", 24)
	p.overview = container.NewBorder(container.NewPadded(container.NewVBox(title, p.status, p.search)), nil, nil, nil,
		container.NewStack(p.grid, container.NewCenter(p.empty)))
	p.photos = components.NewImageGrid(db)
	p.photos.OnImageSelected = func(pic datastore.Picture) {
		editor := components.NewImageEditContainer(db, pic)
		editor.OnUpdate = func(updated datastore.Picture) {
			p.photos.UpdatePicture(updated)
			if p.OnPhotoUpdated != nil {
				p.OnPhotoUpdated()
			}
		}
		editor.OnOpen = func() { p.photos.OpenPhoto(pic) }
		editor.OnCoverChanged = p.Refresh
		p.sidebar.SetContent("Photo details", editor.Layout())
	}
	p.back = widget.NewButtonWithIcon("Back", theme.NavigateBackIcon(), p.openParent)
	p.back.Importance = widget.LowImportance
	edit := widget.NewButtonWithIcon("Album details", theme.InfoIcon(), p.editAlbum)
	edit.Importance = widget.LowImportance
	p.breadcrumbs = container.NewHBox()
	toolbar := container.NewBorder(nil, nil, p.back, edit, container.NewHScroll(p.breadcrumbs))
	p.detail = container.NewBorder(container.NewPadded(toolbar), nil, nil, nil, p.photos.Layout())
	p.view = container.NewStack(p.overview)
	p.content = components.NewInspectorWorkspace(p.view, p.sidebar)
	return p
}
func (p *CollectionPage) openAlbum(album datastore.Album) {
	p.selected = &album
	p.sidebar.Hide()
	p.showAlbumContents(true)
}
func (p *CollectionPage) showOverview() {
	p.selected = nil
	p.sidebar.Hide()
	p.view.Objects = []fyne.CanvasObject{p.overview}
	p.view.Refresh()
}
func (p *CollectionPage) findAlbum(id string) (albumCardData, bool) {
	for _, card := range p.albums {
		if card.album.Id == id {
			return card, true
		}
	}
	return albumCardData{}, false
}
func (p *CollectionPage) openParent() {
	if p.selected != nil {
		if card, ok := p.findAlbum(p.selected.Id); ok {
			if parent, ok := p.findAlbum(card.parentID); ok {
				p.openAlbum(parent.album)
				return
			}
		}
	}
	p.showOverview()
}
func (p *CollectionPage) showAlbumContents(reset bool) {
	if p.selected == nil {
		return
	}
	children := []fyne.CanvasObject{}
	for _, card := range p.albums {
		if card.parentID == p.selected.Id {
			children = append(children, p.makeAlbumCard(card))
		}
	}
	p.photos.SetSubalbums(children)
	p.updateBreadcrumbs()
	if reset {
		p.photos.ShowAlbum(p.selected.Id, p.selected.Name)
	} else {
		p.photos.SetTitle(p.selected.Name)
		p.photos.Reload()
	}
	p.view.Objects = []fyne.CanvasObject{p.detail}
	p.view.Refresh()
}
func (p *CollectionPage) updateBreadcrumbs() {
	root := widget.NewButton("All albums", p.showOverview)
	root.Importance = widget.LowImportance
	objects := []fyne.CanvasObject{root}
	ancestors := []datastore.Album{}
	current, ok := p.findAlbum(p.selected.Id)
	for ok && current.parentID != "" {
		parent, exists := p.findAlbum(current.parentID)
		if !exists {
			break
		}
		ancestors = append(ancestors, parent.album)
		current = parent
	}
	for i := len(ancestors) - 1; i >= 0; i-- {
		album := ancestors[i]
		link := widget.NewButton(album.Name, func() { p.openAlbum(album) })
		link.Importance = widget.LowImportance
		objects = append(objects, widget.NewLabel("/"), link)
	}
	objects = append(objects, widget.NewLabel("/"), widget.NewLabel(p.selected.Name))
	p.breadcrumbs.Objects = objects
	p.breadcrumbs.Refresh()
}
func (p *CollectionPage) editAlbum() {
	if p.selected == nil {
		return
	}
	editor := components.NewCollectionAlbumContainer(p.DataStore, *p.selected)
	editor.OnUpdate = p.Refresh
	p.sidebar.SetContent("Album details", editor.Layout())
}
func (p *CollectionPage) Refresh() {
	p.loaded = true
	p.requestID++
	requestID := p.requestID
	go func() {
		albums, err := p.Albums.GetAll()
		var pictures []datastore.Picture
		if err == nil {
			pictures, err = p.Pictures.GetAll()
		}
		cards := albumCards(albums, pictures)
		fyne.Do(func() {
			if requestID != p.requestID {
				return
			}
			if err != nil {
				p.empty.SetText("Could not load albums: " + err.Error())
				p.empty.Show()
				return
			}
			p.albums = cards
			p.filterAlbums()
			if p.selected != nil {
				if card, exists := p.findAlbum(p.selected.Id); exists {
					album := card.album
					p.selected = &album
					p.showAlbumContents(false)
				} else {
					p.showOverview()
				}
			}
		})
	}()
}

func (p *CollectionPage) filterAlbums() {
	query := strings.ToLower(strings.TrimSpace(p.search.Text))
	objects := []fyne.CanvasObject{}
	roots := 0
	for _, card := range p.albums {
		if card.parentID == "" {
			roots++
		}
		if (query == "" && card.parentID == "") || (query != "" && strings.Contains(strings.ToLower(card.album.Name+" "+card.album.Path), query)) {
			objects = append(objects, p.makeAlbumCard(card))
		}
	}
	p.grid.SetObjects(objects)
	p.status.SetText(fmt.Sprintf("%d albums · Organised by folder", roots))
	if roots != len(p.albums) {
		p.status.SetText(fmt.Sprintf("%d albums · %d including subalbums", roots, len(p.albums)))
	}
	if query != "" {
		if len(objects) == 1 {
			p.status.SetText("1 matching album")
		} else {
			p.status.SetText(fmt.Sprintf("%d matching albums", len(objects)))
		}
	}
	p.empty.Hide()
	if len(objects) == 0 {
		p.empty.SetText("No albums yet. Choose your photo folder in Settings, then scan it from Tasks.")
		if query != "" {
			p.empty.SetText("No matching albums. Try another name or folder.")
		}
		p.empty.Show()
	}
}
func (p *CollectionPage) makeAlbumCard(card albumCardData) fyne.CanvasObject {
	open := func() { p.openAlbum(card.album) }
	cover := components.NewImage(p.DataStore, card.picture, nil, func(datastore.Picture) { open() })
	openButton := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), open)
	openButton.Importance = widget.LowImportance
	nameText := widget.NewLabelWithStyle(card.album.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	nameText.Truncation = fyne.TextTruncateEllipsis
	return container.NewBorder(nil, container.NewBorder(nil, nil, nil, openButton, nameText), nil, nil, cover)
}
func (p *CollectionPage) Layout() fyne.CanvasObject {
	p.photos.Layout()
	if !p.loaded {
		p.Refresh()
	}
	return p.content
}
