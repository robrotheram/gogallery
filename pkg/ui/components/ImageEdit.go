package components

import (
	"fmt"
	"gogallery/pkg/ai"
	"gogallery/pkg/datastore"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ImageEditContainer struct {
	*datastore.DataStore
	selectedPic                             datastore.Picture
	titleEntry, descriptionEntry, tagsEntry *widget.Entry
	OnUpdate                                func(datastore.Picture)
	OnOpen                                  func()
	OnCoverChanged                          func()
}

func NewImageEditContainer(db *datastore.DataStore, pic datastore.Picture) *ImageEditContainer {
	title := widget.NewEntry()
	title.SetText(pic.Name)
	title.SetPlaceHolder("Photo title")
	description := widget.NewMultiLineEntry()
	description.SetText(pic.Caption)
	description.Wrapping = fyne.TextWrapWord
	description.SetMinRowsVisible(3)
	tags := widget.NewEntry()
	tags.SetText(pic.Tags)
	tags.SetPlaceHolder("Landscape, travel, summer")
	return &ImageEditContainer{DataStore: db, selectedPic: pic, titleEntry: title, descriptionEntry: description, tagsEntry: tags}
}
func (c *ImageEditContainer) Layout() fyne.CanvasObject {
	var content *fyne.Container
	status := Hint("")
	status.Hide()
	report := func(message string, failed bool) {
		status.Importance = widget.LowImportance
		if failed {
			status.Importance = widget.DangerImportance
		}
		status.SetText(message)
		status.Show()
		if content != nil {
			content.Refresh()
		}
	}
	visibility := widget.NewSelect([]string{"Private", "Public"}, nil)
	visibility.Selected = "Private"
	if strings.EqualFold(c.selectedPic.Visibility, "PUBLIC") {
		visibility.Selected = "Public"
	}
	save := widget.NewButtonWithIcon("Save changes", theme.DocumentSaveIcon(), nil)
	save.Importance = widget.HighImportance
	save.Disable()
	changed := func(string) { save.Enable(); report("Unsaved changes", false) }
	c.titleEntry.OnChanged = changed
	c.descriptionEntry.OnChanged = changed
	c.tagsEntry.OnChanged = changed
	visibility.OnChanged = changed
	save.OnTapped = func() {
		next := c.selectedPic
		next.Name = strings.TrimSpace(c.titleEntry.Text)
		next.Caption = c.descriptionEntry.Text
		next.Tags = c.tagsEntry.Text
		next.Visibility = strings.ToUpper(visibility.Selected)
		save.Disable()
		c.titleEntry.Disable()
		c.descriptionEntry.Disable()
		c.tagsEntry.Disable()
		visibility.Disable()
		go func() {
			err := c.Pictures.Update(next.Id, next)
			fyne.Do(func() {
				c.titleEntry.Enable()
				c.descriptionEntry.Enable()
				c.tagsEntry.Enable()
				visibility.Enable()
				if err != nil {
					report("Could not save: "+err.Error(), true)
					save.Enable()
					return
				}
				c.selectedPic = next
				report("Changes saved", false)
				if c.OnUpdate != nil {
					c.OnUpdate(next)
				}
			})
		}()
	}
	open := widget.NewButtonWithIcon("View photo", theme.ViewFullScreenIcon(), func() {
		if c.OnOpen != nil {
			c.OnOpen()
		} else {
			ShowPhotoViewer(c.DataStore, []datastore.Picture{c.selectedPic}, c.selectedPic.Id)
		}
	})
	body := container.NewVBox(PhotoPreview(c.DataStore, c.selectedPic, false), open,
		Field("Title", c.titleEntry), Field("Description", c.descriptionEntry), Field("Tags", c.tagsEntry),
		Field("Website visibility", visibility))
	if c.selectedPic.Album != "" {
		cover := widget.NewButtonWithIcon("Use as album cover", theme.MediaPhotoIcon(), nil)
		cover.OnTapped = func() {
			album, err := c.Albums.FindById(c.selectedPic.Album)
			if err == nil {
				album.ProfileId = c.selectedPic.Id
				err = c.Albums.Update(album.Id, album)
			}
			if err != nil {
				report("Could not set cover: "+err.Error(), true)
				return
			}
			report("Album cover updated", false)
			if c.OnCoverChanged != nil {
				c.OnCoverChanged()
			}
		}
		body.Add(cover)
	}
	if ai.IsAIEnabled() {
		generate := widget.NewButtonWithIcon("Suggest title and description", AIIcon(), nil)
		generate.OnTapped = func() {
			generate.Disable()
			generate.SetText("Generating…")
			go func() {
				metadata, err := ai.GenerateMetadata(c.DataStore, c.selectedPic.Id)
				fyne.Do(func() {
					generate.Enable()
					generate.SetText("Suggest metadata")
					if err != nil {
						report(err.Error(), true)
						return
					}
					c.titleEntry.SetText(metadata.Title)
					c.descriptionEntry.SetText(metadata.Description)
					c.tagsEntry.SetText(strings.Join(metadata.Tags, ", "))
					report("Suggestions ready. Review the title, description, and tags, then save.", false)
				})
			}()
		}
		body.Add(generate)
	}
	details := container.NewVBox()
	addDetail := func(name, value string) {
		if value != "" {
			details.Add(Hint(name + "  ·  " + value))
		}
	}
	pic := c.selectedPic
	addDetail("File", filepath.Base(pic.Path))
	addDetail("Camera", pic.Camera)
	addDetail("Lens", pic.LensModel)
	addDetail("Aperture", pic.FStop)
	addDetail("Shutter", pic.ShutterSpeedDisplay())
	addDetail("ISO", pic.ISO)
	addDetail("Focal length", pic.FocalLength)
	addDetail("Dimensions", pic.Dimension)
	if !pic.DateTaken.IsZero() {
		addDetail("Taken", pic.DateTaken.Format("2 Jan 2006, 15:04"))
	}
	if pic.GPSLat != 0 || pic.GPSLng != 0 {
		addDetail("Location", fmt.Sprintf("%.5f, %.5f", pic.GPSLat, pic.GPSLng))
	}
	body.Add(widget.NewAccordion(widget.NewAccordionItem("Photo information", details)))
	content = container.NewBorder(nil, container.NewVBox(status, save), nil, nil, container.NewVScroll(body))
	return content
}
