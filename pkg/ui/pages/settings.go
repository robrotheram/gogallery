package pages

import (
	"fmt"
	"sort"
	"strconv"
	"testingFyne/pkg/config"
	"testingFyne/pkg/datastore"
	"testingFyne/pkg/ui/components"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type SettingsPage struct {
	Title string
	db    *datastore.DataStore
}

func NewSettingsPage(db *datastore.DataStore) *SettingsPage {
	return &SettingsPage{
		Title: "Settings",
		db:    db,
	}
}

func (s *SettingsPage) Layout() fyne.CanvasObject {
	// Sidebar nav
	nav := map[string]fyne.CanvasObject{
		"Gallery":     galleryConfigForm(),
		"Author":      aboutConfigForm(),
		"Deployment":  deployConfigForm(),
		"Application": uiConfigForm(),
		// "Albums":      s.Albums(),
	}

	navItems := make([]string, 0, len(nav))
	for item := range nav {
		navItems = append(navItems, item)
	}

	sort.Slice(navItems, func(i, j int) bool {
		return navItems[i] < navItems[j]
	})

	navList := widget.NewList(
		func() int { return len(navItems) },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(navItems[i])
		},
	)

	// Fyne's widget.List does not support SetMinSize directly. Wrap in a container with MinSize.

	// Content panels for each nav item

	contentStack := container.NewPadded(nav["Gallery"])
	navList.OnSelected = func(id int) {
		contentStack.Objects = []fyne.CanvasObject{nav[navItems[id]]}
		contentStack.Refresh()
	}
	navList.Select(0)

	split := container.NewHSplit(container.NewPadded(navList), contentStack)
	split.Offset = 0.33 // Reduce sidebar width
	return split
}

func galleryConfigForm() fyne.CanvasObject {
	cfg := config.Config.Gallery
	name := widget.NewEntry()
	name.SetText(cfg.Name)
	theme := widget.NewEntry()
	theme.SetText(cfg.Theme)
	imagesPerPage := widget.NewEntry()
	imagesPerPage.SetText(fmt.Sprintf("%d", cfg.ImagesPerPage))

	basePath := widget.NewEntry()
	basePath.SetText(cfg.Basepath)

	destpath := widget.NewEntry()
	destpath.SetText(cfg.Destpath)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Name", Widget: name, HintText: "Gallery name"},
			{Text: "Theme", Widget: theme, HintText: "Path to the theme dir"},
			{Text: "Base Path", Widget: basePath, HintText: "Path to the gallery base directory"},
			{Text: "Destination Path", Widget: destpath, HintText: "Path to the destination directory for images"},
			{Text: "Images Per Page", Widget: imagesPerPage, HintText: "Number of images to display per page"},
		},
		OnCancel: func() {
			name.SetText(cfg.Name)
			theme.SetText(cfg.Theme)
			imagesPerPage.SetText(fmt.Sprintf("%d", cfg.ImagesPerPage))
		},
		OnSubmit: func() {
			cfg.Name = name.Text
			cfg.Theme = theme.Text
			if n, err := strconv.Atoi(imagesPerPage.Text); err == nil {
				cfg.ImagesPerPage = n
			}
			cfg.Save()
		},
	}
	title := components.NewTextEntry("Gallery Settings", 20)
	return container.NewVBox(title, widget.NewSeparator(), form)
}

func aboutConfigForm() fyne.CanvasObject {
	cfg := config.Config.About
	twitter := widget.NewEntry()
	twitter.SetText(cfg.Twitter)
	facebook := widget.NewEntry()
	facebook.SetText(cfg.Facebook)
	email := widget.NewEntry()
	email.SetText(cfg.Email)
	instagram := widget.NewEntry()
	instagram.SetText(cfg.Instagram)
	description := widget.NewEntry()
	description.MultiLine = true
	description.Wrapping = fyne.TextWrapWord
	description.SetMinRowsVisible(5)
	description.SetText(cfg.Description)
	footer := widget.NewEntry()
	footer.SetText(cfg.Footer)
	photographer := widget.NewEntry()
	photographer.SetText(cfg.Photographer)
	profilePhoto := widget.NewEntry()
	profilePhoto.SetText(cfg.ProfilePhoto)
	backgroundPhoto := widget.NewEntry()
	backgroundPhoto.SetText(cfg.BackgroundPhoto)
	blog := widget.NewEntry()
	blog.SetText(cfg.Blog)
	website := widget.NewEntry()
	website.SetText(cfg.Website)
	github := widget.NewEntry()
	github.SetText(cfg.Github)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Twitter", Widget: twitter, HintText: "Twitter handle"},
			{Text: "Facebook", Widget: facebook, HintText: "Facebook page URL"},
			{Text: "Email", Widget: email, HintText: "Contact email address"},
			{Text: "Instagram", Widget: instagram, HintText: "Instagram handle"},
			{Text: "Description", Widget: description, HintText: "Short description of the gallery"},
			{Text: "Footer", Widget: footer, HintText: "Footer text for the gallery"},
			{Text: "Photographer", Widget: photographer, HintText: "Name of the photographer"},
			{Text: "Profile Photo", Widget: profilePhoto, HintText: "Path to the profile photo"},
			{Text: "Background Photo", Widget: backgroundPhoto, HintText: "Path to the background photo"},
			{Text: "Blog", Widget: blog, HintText: "Link to the photographer's blog"},
			{Text: "Website", Widget: website, HintText: "Link to the photographer's website"},
			{Text: "Github", Widget: github, HintText: "Link to the photographer's GitHub profile"},
		},
		OnCancel: func() {
			twitter.SetText(cfg.Twitter)
			facebook.SetText(cfg.Facebook)
			email.SetText(cfg.Email)
			instagram.SetText(cfg.Instagram)
			description.SetText(cfg.Description)
			footer.SetText(cfg.Footer)
			photographer.SetText(cfg.Photographer)
			profilePhoto.SetText(cfg.ProfilePhoto)
			backgroundPhoto.SetText(cfg.BackgroundPhoto)
			blog.SetText(cfg.Blog)
			website.SetText(cfg.Website)
			github.SetText(cfg.Github)
		},
		OnSubmit: func() {
			cfg.Twitter = twitter.Text
			cfg.Facebook = facebook.Text
			cfg.Email = email.Text
			cfg.Instagram = instagram.Text
			cfg.Description = description.Text
			cfg.Footer = footer.Text
			cfg.Photographer = photographer.Text
			cfg.ProfilePhoto = profilePhoto.Text
			cfg.BackgroundPhoto = backgroundPhoto.Text
			cfg.Blog = blog.Text
			cfg.Website = website.Text
			cfg.Github = github.Text
			cfg.Save()
		},
	}
	title := components.NewTextEntry("Author Settings", 20)
	scrollForm := container.NewVScroll(container.NewPadded(form))
	return container.NewBorder(
		container.NewVBox(title, widget.NewSeparator()), // top
		nil,        // bottom
		nil,        // left
		nil,        // right
		scrollForm, // center (fills remaining space)
	)
}

func deployConfigForm() fyne.CanvasObject {
	cfg := config.Config.Deploy
	siteId := widget.NewEntry()
	siteId.SetText(cfg.SiteId)
	authToken := widget.NewEntry()
	authToken.SetText(cfg.AuthToken)
	draft := widget.NewCheck("Draft", func(b bool) { cfg.Draft = b })
	draft.SetChecked(cfg.Draft)
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Site ID", Widget: siteId, HintText: "Your site ID from the deployment service"},
			{Text: "Auth Token", Widget: authToken, HintText: "Your authentication token for the deployment service"},
			{Text: "Draft", Widget: draft, HintText: "Enable draft mode for deployments"},
		},
		OnCancel: func() {
			siteId.SetText(cfg.SiteId)
			authToken.SetText(cfg.AuthToken)
			draft.SetChecked(cfg.Draft)
		},
		OnSubmit: func() {
			cfg.SiteId = siteId.Text
			cfg.AuthToken = authToken.Text
			cfg.Draft = draft.Checked
			cfg.Save()
		},
	}
	title := components.NewTextEntry("Deployment Settings", 20)
	return container.NewVBox(title, widget.NewSeparator(), form)
}

func uiConfigForm() fyne.CanvasObject {
	cfg := &config.Config.UI
	// Theme selection: Light or Dark
	themeOptions := []string{"light", "dark"}
	themeSelect := widget.NewRadioGroup(themeOptions, func(selected string) {
		if selected == "light" {
			cfg.Theme = "light"
		} else {
			cfg.Theme = "dark"
		}
	})
	themeSelect.SetSelected(cfg.Theme)

	notifications := widget.NewCheck("Enable Notifications", func(b bool) {
		cfg.Notification = b
	})
	notifications.SetChecked(cfg.Notification)

	// Preview public checkbox (add PublicPreview to UIConfiguration if not present)
	previewPublic := widget.NewCheck("Preview Public", func(b bool) {
		cfg.Public = b
	})
	if v, ok := any(cfg).(interface{ GetPublicPreview() bool }); ok {
		previewPublic.SetChecked(v.GetPublicPreview())
	} else {
		previewPublic.SetChecked(cfg.Public)
	}

	// Setting the number of images per page
	imagesPerPage := widget.NewEntry()
	imagesPerPage.SetText(fmt.Sprintf("%d", cfg.ImagesPerPage))
	imagesPerPage.OnChanged = func(text string) {
		if n, err := strconv.Atoi(text); err == nil {
			cfg.ImagesPerPage = n
		} else {
			imagesPerPage.SetText(fmt.Sprintf("%d", cfg.ImagesPerPage))
		}
	}

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Theme", Widget: themeSelect, HintText: "Select the application theme requires application restart"},
			{Text: "Notifications", Widget: notifications, HintText: "Enable or disable notifications"},
			{Text: "Public Preview", Widget: previewPublic, HintText: "Enable or disable public preview"},
			{Text: "Images Per Page", Widget: imagesPerPage, HintText: "Number of images to display per page"},
		},
		OnCancel: func() {
			notifications.SetChecked(cfg.Notification)
			previewPublic.SetChecked(cfg.Public)
		},
		OnSubmit: func() {
			cfg.Notification = notifications.Checked
			cfg.Public = previewPublic.Checked
			config.Config.Save()
		},
	}
	title := components.NewTextEntry("Application Settings", 20)
	return container.NewVBox(title, widget.NewSeparator(), form)
}
