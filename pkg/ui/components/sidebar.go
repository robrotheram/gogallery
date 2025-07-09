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
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Sidebar struct {
	*datastore.DataStore
	visible      bool
	selectedPic  datastore.Picture // Reference to the currently selected picture
	titleEntry   *widget.Entry
	captionEntry *widget.Entry
	image        *canvas.Image     // Placeholder for the image to be displayed
	imageStack   *fyne.Container   // Direct reference to the image stack
	container    fyne.CanvasObject // Reference to the sidebar container for refresh
	exifCard     *fyne.Container   // Reference to the EXIF card for updates
	OnClose      func()            // Callback for close button
}

func NewSidebar(db *datastore.DataStore, onClose func()) *Sidebar {
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Enter image title")

	captionEntry := widget.NewMultiLineEntry()
	captionEntry.SetPlaceHolder("Enter image caption")
	captionEntry.Wrapping = fyne.TextWrapWord

	// Create a persistent image widget (bigger size)
	img := canvas.NewImageFromImage(nil) // Start with nil image
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(600, 400))
	imageStack := container.NewStack(img)

	return &Sidebar{
		visible:      false,
		titleEntry:   titleEntry,
		DataStore:    db,
		captionEntry: captionEntry,
		image:        img,
		imageStack:   imageStack,
		OnClose:      onClose,
	}
}

func NewTextEntry(textStr string, size float32) *canvas.Text {
	text := canvas.NewText(textStr, nil)
	text.TextStyle = fyne.TextStyle{Bold: true}
	text.TextSize = size // Larger font size
	return text
}

func (s *Sidebar) Layout() fyne.CanvasObject {
	// Close button
	closeBtn := widget.NewButtonWithIcon("", theme.ContentClearIcon(), s.OnClose)
	closeBtn.Importance = widget.LowImportance
	closeBtn.Alignment = widget.ButtonAlignTrailing
	closeBtnBox := container.NewVBox(
		layout.NewSpacer(),
		closeBtn,
		layout.NewSpacer(),
	)

	titleRow := container.NewHBox(
		NewTextEntry("Image Details", 22),
		layout.NewSpacer(),
		closeBtnBox,
	)

	// EXIF info section (populated in ShowImage)
	form := widget.NewForm(
		widget.NewFormItem("Title", s.titleEntry),
		widget.NewFormItem("Caption", s.captionEntry),
	)
	form.OnSubmit = func() {
		log.Println("Form submitted with title:", s.titleEntry.Text, "and caption:", s.captionEntry.Text)
		// Update the selected picture with new title and caption
		s.selectedPic.Name = s.titleEntry.Text
		s.selectedPic.Caption = s.captionEntry.Text
		if err := s.DataStore.Pictures.Update(s.selectedPic.Id, s.selectedPic); err != nil {
			log.Println("Error updating picture:", err)
		} else {
			log.Println("Picture updated successfully")
			utils.Notify("Update Successful", "Picture details updated successfully")
		}
	}
	s.exifCard = container.NewVBox()

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
				cap, err := ai.GenerateCaption(s.DataStore, s.selectedPic.Id)
				if err != nil {
					return
				}
				fyne.Do(func() {
					s.titleEntry.SetText(cap.Title)
					s.captionEntry.SetText(cap.Caption)
					aiButton.Enable()
					aiButton.SetText("Generate Caption")
				})
			}()

		})
		scrollContent = container.NewVBox(
			titleRow,
			s.imageStack,
			aiButton,
			form,
			widget.NewSeparator(),
			NewTextEntry("EXIF Details", 20),
			s.exifCard,
		)
	} else {
		scrollContent = container.NewVBox(
			titleRow,
			s.imageStack,
			form,
			widget.NewSeparator(),
			NewTextEntry("EXIF Details", 20),
			s.exifCard,
		)
	}

	// Add padding and border
	// padded := container.NewPadded(scrollContent)
	card := widget.NewCard("", "", scrollContent)
	s.container = container.NewVScroll(card)
	s.container.Hide()
	return s.container
}

// Refresh the sidebar UI (call this after ShowImage)
func (s *Sidebar) Refresh() {
	log.Println("Refreshing sidebar")
	if s.visible {
		s.container.Show()
	} else {
		s.container.Hide()
	}
}

func (s *Sidebar) Hide() {
	s.visible = false
	s.container.Hide()
}

func (s *Sidebar) loadImage(pic datastore.Picture) {
	file, err := s.ImageCache.Get(pic.Id, config.JPEG, "small")
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
	s.updateImageStack(data, pic)
}

func (s *Sidebar) updateImageStack(data []byte, pic datastore.Picture) {
	newImg := canvas.NewImageFromReader(bytes.NewReader(data), pic.Name)
	newImg.FillMode = canvas.ImageFillContain
	width := float32(500)
	height := width / pic.AspectRatio
	newImg.SetMinSize(fyne.NewSize(width, height))
	newImg.Resize(fyne.NewSize(width, height))

	if s.imageStack != nil {
		s.imageStack.Objects = []fyne.CanvasObject{newImg}
		s.imageStack.Refresh()
	}
	s.image = newImg

	if s.container != nil {
		s.container.Refresh()
	}
}

// ShowImage sets the sidebar image to the selected picture and refreshes the sidebar
func (s *Sidebar) ShowImage(pic datastore.Picture) {
	log.Println("Showing image:", pic.Id, pic.Name)
	s.selectedPic = pic
	s.visible = true
	s.loadImage(pic)
	s.titleEntry.SetText(pic.Name)
	s.captionEntry.Text = pic.Caption

	// Build EXIF info section
	exifLabels := []fyne.CanvasObject{
		widget.NewLabel("Camera: " + pic.Camera),
		widget.NewLabel("Lens: " + pic.LensModel),
		widget.NewLabel("F-Stop: " + pic.FStop),
		widget.NewLabel("Shutter: " + pic.ShutterSpeed),
		widget.NewLabel("ISO: " + pic.ISO),
		widget.NewLabel("Focal Length: " + pic.FocalLength),
		widget.NewLabel("Date Taken: " + pic.DateTaken.Format("2006-01-02 15:04:05")),
		widget.NewLabel("Dimensions: " + pic.Dimension),
		widget.NewLabel("Aspect Ratio: " + fmt.Sprintf("%.2f", pic.AspectRatio)),
		widget.NewLabel("GPS: " + fmt.Sprintf("%.6f, %.6f", pic.GPSLat, pic.GPSLng)),
	}
	if s.exifCard != nil {
		s.exifCard.Objects = exifLabels
		s.exifCard.Refresh()
	}
	if s.container != nil {
		s.container.Refresh()
	}
	s.Refresh()
}
