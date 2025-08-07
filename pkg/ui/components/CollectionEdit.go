package components

import (
	"bytes"
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"io"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type CollectionAlbumContainer struct {
	*datastore.DataStore
	selectedAlbum datastore.Album // Reference to the currently selected album
	titleEntry    *widget.Entry
	image         *canvas.Image     // Placeholder for the image to be displayed
	imageStack    *fyne.Container   // Direct reference to the image stack
	container     fyne.CanvasObject // Reference to the sidebar container for refresh
	OnUpdate      func()            // Callback for album update
}

func NewCollectionAlbumContainer(db *datastore.DataStore, selectedAlbum datastore.Album) *CollectionAlbumContainer {
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Enter album title")

	// Use a placeholder image instead of nil to avoid layout issues on Windows
	placeholder := canvas.NewRectangle(nil)
	placeholder.SetMinSize(fyne.NewSize(600, 400))
	img := canvas.NewImageFromImage(nil)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(600, 400)) // More reasonable default
	imageStack := container.NewStack(placeholder)

	return &CollectionAlbumContainer{
		DataStore:     db,
		selectedAlbum: selectedAlbum,
		titleEntry:    titleEntry,
		image:         img,
		imageStack:    imageStack,
	}
}

func (c *CollectionAlbumContainer) loadImage(pic datastore.Picture) {

	file, err := c.ImageCache.Get(pic.Id, config.JPEG, "small")
	if err != nil {
		log.Println("Error loading image from cache:", err)
		return
	}
	data, err := io.ReadAll(file)
	if err != nil {
		log.Println("Error reading image file:", err)
		return
	}
	log.Printf("[Sidebar] Loaded image bytes: %d for %s", len(data), pic.Name)
	if len(data) < 16 {
		log.Println("[Sidebar] Image data too small or empty, not displaying.")
		return
	}

	newImg := canvas.NewImageFromReader(bytes.NewReader(data), pic.Name)
	newImg.FillMode = canvas.ImageFillContain
	width := float32(500)
	if pic.AspectRatio > 0 {
		height := width / pic.AspectRatio
		newImg.SetMinSize(fyne.NewSize(width, height))
		newImg.Resize(fyne.NewSize(width, height))
	} else {
		newImg.SetMinSize(fyne.NewSize(width, 300))
		newImg.Resize(fyne.NewSize(width, 300))
	}

	if c.imageStack != nil {
		c.imageStack.Objects = []fyne.CanvasObject{newImg}
		c.imageStack.Refresh()
	}
	c.image = newImg
	c.imageStack.Refresh()
}

func (c *CollectionAlbumContainer) Layout() fyne.CanvasObject {

	if pic, err := c.DataStore.Pictures.FindById(c.selectedAlbum.ProfileId); err == nil {
		c.loadImage(pic)
	}
	c.titleEntry.SetText(c.selectedAlbum.Name)

	var imageOptions []string
	imageMap := make(map[string]datastore.Picture)
	selectedImage := c.selectedAlbum.ProfileId
	if pics, err := c.Pictures.FindByField("album", c.selectedAlbum.Id); err == nil {
		imageOptions = make([]string, len(pics))
		for i, pic := range pics {
			imageOptions[i] = pic.Name
			imageMap[pic.Name] = pic
			if pic.Id == c.selectedAlbum.ProfileId {
				selectedImage = pic.Name // Set the selected image name
			}
		}
	}

	imageSelect := widget.NewSelect(imageOptions, func(selected string) {
		selectedImage = imageMap[selected].Id
		if pic, exists := imageMap[selected]; exists {
			c.loadImage(pic)
		}
	})
	imageSelect.Selected = selectedImage

	// EXIF info section (populated in ShowImage)
	form := widget.NewForm(
		widget.NewFormItem("Title", c.titleEntry),
		widget.NewFormItem("Feature Image", imageSelect),
	)

	form.OnSubmit = func() {
		c.selectedAlbum.Name = c.titleEntry.Text
		c.selectedAlbum.ProfileId = selectedImage
		if err := c.DataStore.Albums.Update(c.selectedAlbum.Id, c.selectedAlbum); err != nil {
			log.Println("Error updating album name:", err)
			return
		}
		log.Printf("Updated album name to: %s", c.selectedAlbum.Name)
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Album Updated",
			Content: "The album details have been saved successfully.",
		})
		if c.OnUpdate != nil {
			c.OnUpdate() // Call the update callback if set
		}
	}

	scrollContent := container.NewVBox(
		c.imageStack,
		form,
	)

	c.container = container.NewVScroll(scrollContent)
	return c.container
}
