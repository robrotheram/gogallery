package components

import (
	"bytes"
	"fmt"
	"gogallery/pkg/ai"
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"gogallery/pkg/ui/utils"
	"io"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ImageEditContainer struct {
	*datastore.DataStore
	selectedPic  datastore.Picture // Reference to the currently selected picture
	titleEntry   *widget.Entry
	captionEntry *widget.Entry
	image        *canvas.Image     // Placeholder for the image to be displayed
	imageStack   *fyne.Container   // Direct reference to the image stack
	container    fyne.CanvasObject // Reference to the sidebar container for refresh
	exifCard     *fyne.Container   // Reference to the EXIF card for updates
}

func NewImageEditContainer(db *datastore.DataStore, selectedPic datastore.Picture) *ImageEditContainer {
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Enter image title")

	captionEntry := widget.NewMultiLineEntry()
	captionEntry.SetPlaceHolder("Enter image caption")
	captionEntry.Wrapping = fyne.TextWrapWord

	// Use a placeholder image instead of nil to avoid layout issues on Windows
	placeholder := canvas.NewRectangle(nil)
	placeholder.SetMinSize(fyne.NewSize(600, 400))
	img := canvas.NewImageFromImage(nil)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(600, 400)) // More reasonable default
	imageStack := container.NewStack(placeholder)

	return &ImageEditContainer{
		DataStore:    db,
		selectedPic:  selectedPic,
		titleEntry:   titleEntry,
		captionEntry: captionEntry,
		image:        img,
		imageStack:   imageStack,
	}
}

func (c *ImageEditContainer) loadImage() {
	file, err := c.ImageCache.Get(c.selectedPic.Id, config.JPEG, "small")
	if err != nil {
		log.Println("Error loading image from cache:", err)
		return
	}
	data, err := io.ReadAll(file)
	if err != nil {
		log.Println("Error reading image file:", err)
		return
	}
	log.Printf("[Sidebar] Loaded image bytes: %d for %s", len(data), c.selectedPic.Name)
	if len(data) < 16 {
		log.Println("[Sidebar] Image data too small or empty, not displaying.")
		return
	}

	newImg := canvas.NewImageFromReader(bytes.NewReader(data), c.selectedPic.Name)
	newImg.FillMode = canvas.ImageFillContain
	width := float32(500)
	if c.selectedPic.AspectRatio > 0 {
		height := width / c.selectedPic.AspectRatio
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
}

func (c *ImageEditContainer) Layout() fyne.CanvasObject {
	c.loadImage()

	c.titleEntry.SetText(c.selectedPic.Name)
	c.captionEntry.Text = c.selectedPic.Caption

	// EXIF info section (populated in ShowImage)
	form := widget.NewForm(
		widget.NewFormItem("Title", c.titleEntry),
		widget.NewFormItem("Caption", c.captionEntry),
	)

	form.OnSubmit = func() {
		log.Println("Form submitted with title:", c.titleEntry.Text, "and caption:", c.captionEntry.Text)
		// Update the selected picture with new title and caption
		c.selectedPic.Name = c.titleEntry.Text
		c.selectedPic.Caption = c.captionEntry.Text
		if err := c.DataStore.Pictures.Update(c.selectedPic.Id, c.selectedPic); err != nil {
			log.Println("Error updating picture:", err)
		} else {
			log.Println("Picture updated successfully")
			utils.Notify("Update Successful", "Picture details updated successfully")
		}
	}
	c.exifCard = container.NewVBox(
		widget.NewLabel("Camera: "+c.selectedPic.Camera),
		widget.NewLabel("Lens: "+c.selectedPic.LensModel),
		widget.NewLabel("F-Stop: "+c.selectedPic.FStop),
		widget.NewLabel("Shutter: "+c.selectedPic.ShutterSpeed),
		widget.NewLabel("ISO: "+c.selectedPic.ISO),
		widget.NewLabel("Focal Length: "+c.selectedPic.FocalLength),
		widget.NewLabel("Date Taken: "+c.selectedPic.DateTaken.Format("2006-01-02 15:04:05")),
		widget.NewLabel("Dimensions: "+c.selectedPic.Dimension),
		widget.NewLabel("Aspect Ratio: "+fmt.Sprintf("%.2f", c.selectedPic.AspectRatio)),
		widget.NewLabel("GPS: "+fmt.Sprintf("%.6f, %.6f", c.selectedPic.GPSLat, c.selectedPic.GPSLng)),
	)

	//AI button
	var scrollContent *fyne.Container
	if ai.IsAi() {
		var aiButton *widget.Button
		aiButton = widget.NewButtonWithIcon("Generate Caption", theme.ContentAddIcon(), func() {
			go func() {
				fyne.Do(func() {
					aiButton.Disable()
					aiButton.SetText("Generating...")
				})
				cap, err := ai.GenerateCaption(c.DataStore, c.selectedPic.Id)
				if err != nil {
					return
				}
				fyne.Do(func() {
					c.titleEntry.SetText(cap.Title)
					c.captionEntry.SetText(cap.Caption)
					aiButton.Enable()
					aiButton.SetText("Generate Caption")
				})
			}()

		})
		scrollContent = container.NewVBox(
			c.imageStack,
			aiButton,
			form,
			widget.NewSeparator(),
			NewTextEntry("EXIF Details", 20),
			c.exifCard,
		)
	} else {
		scrollContent = container.NewVBox(
			c.imageStack,
			form,
			widget.NewSeparator(),
			NewTextEntry("EXIF Details", 20),
			c.exifCard,
		)
	}

	c.container = container.NewVScroll(scrollContent)
	return c.container
}
