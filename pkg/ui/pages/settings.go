package pages

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gogallery/pkg/ai"
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"gogallery/pkg/preview"
	"gogallery/pkg/ui/components"
	"gogallery/pkg/ui/utils"
)

type SettingsPage struct {
	Title  string
	db     *datastore.DataStore
	server *preview.Server
}

func NewSettingsPage(db *datastore.DataStore, servers ...*preview.Server) *SettingsPage {
	page := &SettingsPage{
		Title: "Settings",
		db:    db,
	}
	if len(servers) > 0 {
		page.server = servers[0]
	}
	return page
}

type settingsNotice struct {
	container    *fyne.Container
	icon         *widget.Icon
	title        *widget.Label
	message      *widget.Label
	timerMu      sync.Mutex
	dismissTimer *time.Timer
}

const settingsNoticeDuration = 20 * time.Second

func newSettingsNotice() *settingsNotice {
	notice := &settingsNotice{
		icon:    widget.NewIcon(theme.NewThemedResource(theme.Icon(theme.IconNameConfirm))),
		title:   widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		message: widget.NewLabel(""),
	}
	notice.message.Wrapping = fyne.TextWrapWord

	dismiss := widget.NewButtonWithIcon("", theme.Icon(theme.IconNameCancel), notice.Hide)
	dismiss.Importance = widget.LowImportance
	body := container.NewBorder(
		nil,
		nil,
		notice.icon,
		container.NewCenter(dismiss),
		container.NewVBox(notice.title, notice.message),
	)
	notice.container = container.NewStack(widget.NewCard("", "", body))
	notice.Hide()
	return notice
}

func (n *settingsNotice) Hide() {
	n.timerMu.Lock()
	if n.dismissTimer != nil {
		n.dismissTimer.Stop()
		n.dismissTimer = nil
	}
	n.timerMu.Unlock()
	n.container.Hide()
}

func (n *settingsNotice) scheduleDismiss() {
	n.timerMu.Lock()
	if n.dismissTimer != nil {
		n.dismissTimer.Stop()
	}

	var timer *time.Timer
	timer = time.AfterFunc(settingsNoticeDuration, func() {
		fyne.Do(func() {
			n.timerMu.Lock()
			if n.dismissTimer != timer {
				n.timerMu.Unlock()
				return
			}
			n.dismissTimer = nil
			n.timerMu.Unlock()
			n.container.Hide()
		})
	})
	n.dismissTimer = timer
	n.timerMu.Unlock()
}

func (n *settingsNotice) ShowSuccess(section string) {
	n.icon.SetResource(theme.NewThemedResource(theme.Icon(theme.IconNameConfirm)))
	n.title.Importance = widget.MediumImportance
	n.title.SetText("Changes saved")
	n.message.SetText(section + " settings are up to date.")
	n.container.Show()
	n.scheduleDismiss()
}

func (n *settingsNotice) ShowError(err error) {
	n.icon.SetResource(theme.NewErrorThemedResource(theme.Icon(theme.IconNameError)))
	n.title.Importance = widget.DangerImportance
	n.title.SetText("Could not save settings")
	n.message.SetText(err.Error())
	n.container.Show()
	n.scheduleDismiss()
}

func showSettingsSaveResult(notice *settingsNotice, section string, err error) {
	if err != nil {
		notice.ShowError(err)
		utils.Notify("Settings not saved", err.Error())
		return
	}

	message := section + " settings saved"
	notice.ShowSuccess(section)
	utils.Notify("Settings saved", message)
}

func (s *SettingsPage) Layout() fyne.CanvasObject {
	// Sidebar nav
	nav := map[string]fyne.CanvasObject{
		"Gallery":     galleryConfigForm(),
		"Author":      aboutConfigForm(),
		"Deployment":  deployConfigForm(),
		"Application": uiConfigForm(s.server),
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
	status := newSettingsNotice()
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
		SubmitText: "Save",
		CancelText: "Reset",
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
			basePath.SetText(cfg.Basepath)
			destpath.SetText(cfg.Destpath)
			imagesPerPage.SetText(fmt.Sprintf("%d", cfg.ImagesPerPage))
			status.Hide()
		},
		OnSubmit: func() {
			n, err := strconv.Atoi(strings.TrimSpace(imagesPerPage.Text))
			if err != nil || n <= 0 {
				showSettingsSaveResult(status, "Gallery", fmt.Errorf("images per page must be a positive number"))
				return
			}
			next := cfg
			next.Name = strings.TrimSpace(name.Text)
			next.Theme = strings.TrimSpace(theme.Text)
			next.Basepath = strings.TrimSpace(basePath.Text)
			next.Destpath = strings.TrimSpace(destpath.Text)
			next.ImagesPerPage = n
			err = next.Save()
			if err == nil {
				cfg = next
			}
			showSettingsSaveResult(status, "Gallery", err)
		},
	}
	title := components.NewTextEntry("Gallery Settings", 20)
	return container.NewVBox(title, widget.NewSeparator(), status.container, form)
}

func aboutConfigForm() fyne.CanvasObject {
	cfg := config.Config.About
	status := newSettingsNotice()
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
		SubmitText: "Save",
		CancelText: "Reset",
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
			status.Hide()
		},
		OnSubmit: func() {
			next := cfg
			next.Twitter = twitter.Text
			next.Facebook = facebook.Text
			next.Email = email.Text
			next.Instagram = instagram.Text
			next.Description = description.Text
			next.Footer = footer.Text
			next.Photographer = photographer.Text
			next.ProfilePhoto = profilePhoto.Text
			next.BackgroundPhoto = backgroundPhoto.Text
			next.Blog = blog.Text
			next.Website = website.Text
			next.Github = github.Text
			err := next.Save()
			if err == nil {
				cfg = next
			}
			showSettingsSaveResult(status, "Author", err)
		},
	}
	title := components.NewTextEntry("Author Settings", 20)
	scrollForm := container.NewVScroll(container.NewPadded(form))
	return container.NewBorder(
		container.NewVBox(title, widget.NewSeparator(), status.container), // top
		nil,        // bottom
		nil,        // left
		nil,        // right
		scrollForm, // center (fills remaining space)
	)
}

func deployConfigForm() fyne.CanvasObject {
	cfg := config.Config.Deploy
	status := newSettingsNotice()
	siteId := widget.NewEntry()
	siteId.SetText(cfg.SiteId)
	authToken := widget.NewPasswordEntry()
	authToken.SetText(cfg.AuthToken)
	draft := widget.NewCheck("Draft", nil)
	draft.SetChecked(cfg.Draft)
	form := &widget.Form{
		SubmitText: "Save",
		CancelText: "Reset",
		Items: []*widget.FormItem{
			{Text: "Site ID", Widget: siteId, HintText: "Your site ID from the deployment service"},
			{Text: "Auth Token", Widget: authToken, HintText: "Your authentication token for the deployment service"},
			{Text: "Draft", Widget: draft, HintText: "Enable draft mode for deployments"},
		},
		OnCancel: func() {
			siteId.SetText(cfg.SiteId)
			authToken.SetText(cfg.AuthToken)
			draft.SetChecked(cfg.Draft)
			status.Hide()
		},
		OnSubmit: func() {
			next := cfg
			next.SiteId = strings.TrimSpace(siteId.Text)
			next.AuthToken = strings.TrimSpace(authToken.Text)
			next.Draft = draft.Checked
			err := next.Save()
			if err == nil {
				cfg = next
			}
			showSettingsSaveResult(status, "Deployment", err)
		},
	}
	title := components.NewTextEntry("Deployment Settings", 20)
	return container.NewVBox(title, widget.NewSeparator(), status.container, form)
}

func uiConfigForm(server *preview.Server) fyne.CanvasObject {
	cfg := config.Config.UI
	status := newSettingsNotice()
	// Theme selection: Light or Dark
	themeOptions := []string{"light", "dark"}
	themeSelect := widget.NewRadioGroup(themeOptions, nil)
	themeSelect.SetSelected(cfg.Theme)

	notifications := widget.NewCheck("Enable Notifications", nil)
	notifications.SetChecked(cfg.Notification)

	previewPublic := widget.NewCheck("Preview Public", nil)
	previewPublic.SetChecked(cfg.Public)
	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.SetPlaceHolder("Enter Gemini API Key")
	apiKeyEntry.SetText(cfg.GeminiApiKey)

	// Setting the number of images per page
	imagesPerPage := widget.NewEntry()
	imagesPerPage.SetText(fmt.Sprintf("%d", cfg.ImagesPerPage))

	form := &widget.Form{
		SubmitText: "Save",
		CancelText: "Reset",
		Items: []*widget.FormItem{
			{Text: "Theme", Widget: themeSelect, HintText: "Theme changes apply after restarting the application"},
			{Text: "Notifications", Widget: notifications, HintText: "Enable or disable notifications"},
			{Text: "Public Preview", Widget: previewPublic, HintText: "Allow preview access from other devices on your network"},
			{Text: "Images Per Page", Widget: imagesPerPage, HintText: "Number of images to display per page"},
			{Text: "Gemini API Key", Widget: apiKeyEntry, HintText: "Enter your Gemini API key to enable AI features"},
		},
		OnCancel: func() {
			themeSelect.SetSelected(cfg.Theme)
			notifications.SetChecked(cfg.Notification)
			previewPublic.SetChecked(cfg.Public)
			imagesPerPage.SetText(fmt.Sprintf("%d", cfg.ImagesPerPage))
			apiKeyEntry.SetText(cfg.GeminiApiKey)
			status.Hide()
		},
		OnSubmit: func() {
			n, err := strconv.Atoi(strings.TrimSpace(imagesPerPage.Text))
			if err != nil || n <= 0 {
				showSettingsSaveResult(status, "Application", fmt.Errorf("images per page must be a positive number"))
				return
			}
			if themeSelect.Selected == "" {
				showSettingsSaveResult(status, "Application", fmt.Errorf("select a light or dark theme"))
				return
			}

			next := cfg
			next.Theme = themeSelect.Selected
			next.Notification = notifications.Checked
			next.Public = previewPublic.Checked
			next.ImagesPerPage = n
			next.GeminiApiKey = strings.TrimSpace(apiKeyEntry.Text)
			previewAccessChanged := next.Public != cfg.Public
			candidate := *config.Config
			candidate.UI = next
			err = candidate.Save()
			if err == nil {
				cfg = next
				if previewAccessChanged && server != nil {
					go func() {
						if stopErr := server.Stop(); stopErr != nil {
							log.Printf("Could not restart preview access mode: %v", stopErr)
						}
					}()
				}
			}
			showSettingsSaveResult(status, "Application", err)
			if err == nil && next.GeminiApiKey == "" {
				ai.ClearGeminiClient()
			} else if err == nil {
				if _, registerErr := ai.RegisterGeminiClient(); registerErr != nil {
					log.Printf("Could not configure Gemini: %v", registerErr)
				}
			}
		},
	}
	title := components.NewTextEntry("Application Settings", 20)
	return container.NewVBox(title, widget.NewSeparator(), status.container, form)
}
