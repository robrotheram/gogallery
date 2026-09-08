package components

import (
	"fmt"
	"gogallery/pkg/ai"
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"gogallery/pkg/ui/utils"
	"image"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ImageEditContainer struct {
	*datastore.DataStore
	selectedPic      datastore.Picture // Reference to the currently selected picture
	titleEntry       *widget.Entry
	descriptionEntry *widget.Entry
	tagsEntry        *widget.Entry
	imageStack       *fyne.Container   // Direct reference to the image stack
	container        fyne.CanvasObject // Reference to the sidebar container for refresh
	exifCard         *fyne.Container   // Reference to the EXIF card for updates
}

func NewImageEditContainer(db *datastore.DataStore, selectedPic datastore.Picture) *ImageEditContainer {
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Enter image title")

	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("Describe the image")
	descriptionEntry.Wrapping = fyne.TextWrapWord

	tagsEntry := widget.NewEntry()
	tagsEntry.SetPlaceHolder("landscape, golden hour, mountains")

	// Use a placeholder image instead of nil to avoid layout issues on Windows
	placeholder := canvas.NewRectangle(nil)
	placeholder.SetMinSize(fyne.NewSize(360, 240))
	imageStack := container.NewStack(placeholder)

	return &ImageEditContainer{
		DataStore:        db,
		selectedPic:      selectedPic,
		titleEntry:       titleEntry,
		descriptionEntry: descriptionEntry,
		tagsEntry:        tagsEntry,
		imageStack:       imageStack,
	}
}

func (c *ImageEditContainer) loadImage() {
	go func() {
		file, err := c.ImageCache.Get(c.selectedPic.Id, config.JPEG, "small")
		if err != nil {
			log.Println("Error loading image from cache:", err)
			return
		}
		decoded, _, err := image.Decode(file)
		_ = file.Close()
		if err != nil {
			log.Println("Error decoding cached image:", err)
			return
		}

		fyne.Do(func() {
			newImg := canvas.NewImageFromImage(decoded)
			newImg.FillMode = canvas.ImageFillContain
			width := float32(360)
			height := float32(240)
			if c.selectedPic.AspectRatio > 0 {
				height = width / c.selectedPic.AspectRatio
			}
			newImg.SetMinSize(fyne.NewSize(width, height))
			newImg.Resize(fyne.NewSize(width, height))

			if c.imageStack != nil {
				c.imageStack.Objects = []fyne.CanvasObject{newImg}
				c.imageStack.Refresh()
			}
		})
	}()
}

func (c *ImageEditContainer) Layout() fyne.CanvasObject {
	c.loadImage()

	c.titleEntry.SetText(c.selectedPic.Name)
	c.descriptionEntry.SetText(c.selectedPic.Caption)
	c.tagsEntry.SetText(c.selectedPic.Tags)

	// EXIF info section (populated in ShowImage)
	form := widget.NewForm(
		widget.NewFormItem("Title", c.titleEntry),
		widget.NewFormItem("Description", c.descriptionEntry),
		widget.NewFormItem("Tags", c.tagsEntry),
	)

	form.OnSubmit = func() {
		log.Println("Saving metadata for picture:", c.selectedPic.Id)
		c.selectedPic.Name = c.titleEntry.Text
		c.selectedPic.Caption = c.descriptionEntry.Text
		c.selectedPic.Tags = c.tagsEntry.Text
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
	if ai.IsAIEnabled() {
		var aiButton *widget.Button
		aiButton = widget.NewButtonWithIcon("Generate metadata", theme.ContentAddIcon(), func() {
			go func() {
				fyne.Do(func() {
					aiButton.Disable()
					aiButton.SetText("Generating...")
				})
				metadata, err := ai.GenerateMetadata(c.DataStore, c.selectedPic.Id)
				if err != nil {
					fyne.Do(func() {
						aiButton.Enable()
						aiButton.SetText("Generate metadata")
						utils.Notify("AI metadata failed", err.Error())
					})
					return
				}
				fyne.Do(func() {
					c.selectedPic.Name = metadata.Title
					c.selectedPic.Caption = metadata.Description
					c.selectedPic.Tags = strings.Join(metadata.Tags, ", ")
					c.titleEntry.SetText(metadata.Title)
					c.descriptionEntry.SetText(metadata.Description)
					c.tagsEntry.SetText(c.selectedPic.Tags)
					aiButton.Enable()
					aiButton.SetText("Generate metadata")
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
