package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gogallery/pkg/datastore"
	"strings"
)

type CollectionAlbumContainer struct {
	*datastore.DataStore
	selectedAlbum datastore.Album
	titleEntry    *widget.Entry
	OnUpdate      func()
}

func NewCollectionAlbumContainer(db *datastore.DataStore, album datastore.Album) *CollectionAlbumContainer {
	title := widget.NewEntry()
	title.SetText(album.Name)
	title.SetPlaceHolder("Album title")
	return &CollectionAlbumContainer{DataStore: db, selectedAlbum: album, titleEntry: title}
}
func (c *CollectionAlbumContainer) Layout() fyne.CanvasObject {
	pic := datastore.Picture{}
	if c.selectedAlbum.ProfileId != "" {
		pic, _ = c.Pictures.FindByID(c.selectedAlbum.ProfileId)
	}
	if pic.Id == "" {
		pic, _ = c.Pictures.FindLatestInAlbum(c.selectedAlbum.Id)
	}
	var content *fyne.Container
	status := Hint("")
	status.Hide()
	save := widget.NewButtonWithIcon("Save album", theme.DocumentSaveIcon(), nil)
	save.Importance = widget.HighImportance
	save.Disable()
	c.titleEntry.OnChanged = func(string) {
		save.Enable()
		status.SetText("Unsaved changes")
		status.Show()
		if content != nil {
			content.Refresh()
		}
	}
	save.OnTapped = func() {
		name := strings.TrimSpace(c.titleEntry.Text)
		if name == "" {
			status.SetText("Give this album a title.")
			status.Show()
			if content != nil {
				content.Refresh()
			}
			return
		}
		// Read the latest album so editing its title cannot overwrite a cover that
		// was changed from another photo inspector.
		album, err := c.Albums.FindById(c.selectedAlbum.Id)
		if err == nil {
			album.Name = name
			err = c.Albums.Update(album.Id, album)
		}
		if err != nil {
			status.SetText("Could not save: " + err.Error())
			status.Show()
			if content != nil {
				content.Refresh()
			}
			return
		}
		c.selectedAlbum = album
		save.Disable()
		status.SetText("Album saved")
		status.Show()
		if content != nil {
			content.Refresh()
		}
		if c.OnUpdate != nil {
			c.OnUpdate()
		}
	}
	body := container.NewVBox(PhotoPreview(c.DataStore, pic, false), Field("Album title", c.titleEntry),
		Hint("To change the cover, select a photo in this album and choose “Use as album cover”."),
		widget.NewSeparator(), Field("Folder", Hint(c.selectedAlbum.Path)), Hint("Changing the title keeps your folder name unchanged."))
	content = container.NewBorder(nil, container.NewVBox(status, save), nil, nil, container.NewVScroll(body))
	if datastore.IsAlbumInBlacklist(c.selectedAlbum.Name) {
		body.Add(Hint("This album is excluded from the website by your gallery settings."))
	}
	return content
}
