package pages

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
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
	Title   string
	db      *datastore.DataStore
	server  *preview.Server
	content fyne.CanvasObject
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
	onChanged    func()
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
	if n.onChanged != nil {
		n.onChanged()
	}
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
			if n.onChanged != nil {
				n.onChanged()
			}
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
	if n.onChanged != nil {
		n.onChanged()
	}
	n.scheduleDismiss()
}

func (n *settingsNotice) ShowError(err error) {
	n.icon.SetResource(theme.NewErrorThemedResource(theme.Icon(theme.IconNameError)))
	n.title.Importance = widget.DangerImportance
	n.title.SetText("Could not save settings")
	n.message.SetText(err.Error())
	n.container.Show()
	if n.onChanged != nil {
		n.onChanged()
	}
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
	if s.content != nil {
		return s.content
	}
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Library", theme.FolderIcon(), galleryConfigForm()),
		container.NewTabItemWithIcon("Author", theme.AccountIcon(), aboutConfigForm()),
		container.NewTabItemWithIcon("Application", theme.SettingsIcon(), uiConfigForm(s.server)),
		container.NewTabItemWithIcon("Publishing", theme.UploadIcon(), deployConfigForm()),
	)
	heading := components.NewTextEntry("Settings", 28)
	page := container.NewBorder(container.New(layout.NewCustomPaddedLayout(0, 12, 8, 0), heading), nil, nil, nil, tabs)
	s.content = components.WorkspaceFrame(page, 1400)
	return s.content
}

// Each section has one scroller and a fixed action bar. Long paths and the
// author form cannot push the save controls outside the window.
func settingsPanel(title, description string, status *settingsNotice, form *widget.Form, groups ...settingsGroup) fyne.CanvasObject {
	heading := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	heading.SizeName = theme.SizeNameSubHeadingText
	heading.Wrapping = fyne.TextWrapWord
	introduction := container.NewVBox(heading, components.Hint(description))
	sections := []fyne.CanvasObject{}
	for _, group := range groups {
		sections = append(sections, group.content())
	}
	save := widget.NewButtonWithIcon("Save changes", theme.DocumentSaveIcon(), form.OnSubmit)
	save.Importance = widget.HighImportance
	reset := widget.NewButton("Reset changes", form.OnCancel)
	reset.Importance = widget.LowImportance
	actions := container.NewHBox(layout.NewSpacer(), reset, save)
	footer := container.NewVBox(widget.NewSeparator(), status.container,
		container.New(layout.NewCustomPaddedLayout(12, 0, 0, 0), actions))
	panel := container.NewBorder(nil, container.NewStack(canvas.NewRectangle(theme.Color(theme.ColorNameBackground)), footer),
		nil, nil, newSettingsBody(introduction, sections...))
	status.onChanged = panel.Refresh
	return panel
}

func folderField(entry *widget.Entry) fyne.CanvasObject {
	browse := widget.NewButtonWithIcon("Choose…", theme.FolderOpenIcon(), func() {
		app := fyne.CurrentApp()
		if app == nil || len(app.Driver().AllWindows()) == 0 {
			return
		}
		window := app.Driver().AllWindows()[0]
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			if uri != nil {
				entry.SetText(uri.Path())
			}
		}, window)
	})
	return container.NewBorder(nil, nil, nil, browse, entry)
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
			{Text: "Theme", Widget: theme, HintText: "Built-in theme name or path to a custom website theme"},
			{Text: "Photo folder", Widget: folderField(basePath), HintText: "Albums are the folders inside this directory. Scan it from Tasks after changing it."},
			{Text: "Website output", Widget: folderField(destpath), HintText: "Generated website files. Choose a folder separate from your photos."},
			{Text: "Images Per Page", Widget: imagesPerPage, HintText: "Images per page on the generated website, where supported by the theme. The desktop gallery adapts to the window size."},
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
	return settingsPanel("Library & website", "Choose where your photos live and how your website is generated.", status, form,
		settingsGroup{"Gallery", []*widget.FormItem{form.Items[0], form.Items[1], form.Items[4]}},
		settingsGroup{"Folders", form.Items[2:4]})
}

func aboutConfigForm() fyne.CanvasObject {
	cfg := config.Config.About
	status := newSettingsNotice()
	twitter := widget.NewEntry()
	twitter.SetText(cfg.Twitter)
	instagram := widget.NewEntry()
	instagram.SetText(cfg.Instagram)
	description := widget.NewEntry()
	description.MultiLine = true
	description.Wrapping = fyne.TextWrapWord
	description.SetMinRowsVisible(5)
	description.SetText(cfg.Description)
	photographer := widget.NewEntry()
	photographer.SetText(cfg.Photographer)
	profilePhoto := widget.NewEntry()
	profilePhoto.SetText(cfg.ProfilePhoto)
	website := widget.NewEntry()
	website.SetText(cfg.WebsiteURL())
	github := widget.NewEntry()
	github.SetText(cfg.Github)

	form := &widget.Form{
		SubmitText: "Save",
		CancelText: "Reset",
		Items: []*widget.FormItem{
			{Text: "Photographer", Widget: photographer, HintText: "Name of the photographer"},
			{Text: "Description", Widget: description, HintText: "Biography shown on your About page"},
			{Text: "Profile Photo", Widget: profilePhoto, HintText: "Path to the profile photo"},
			{Text: "Website", Widget: website, HintText: "Link to your website or blog"},
			{Text: "Instagram", Widget: instagram, HintText: "Instagram handle"},
			{Text: "X.com", Widget: twitter, HintText: "X handle (formerly Twitter)"},
			{Text: "Github", Widget: github, HintText: "Link to the photographer's GitHub profile"},
		},
		OnCancel: func() {
			twitter.SetText(cfg.Twitter)
			instagram.SetText(cfg.Instagram)
			description.SetText(cfg.Description)
			photographer.SetText(cfg.Photographer)
			profilePhoto.SetText(cfg.ProfilePhoto)
			website.SetText(cfg.WebsiteURL())
			github.SetText(cfg.Github)
			status.Hide()
		},
		OnSubmit: func() {
			next := cfg
			next.Twitter = twitter.Text
			next.Instagram = instagram.Text
			next.Description = description.Text
			next.Photographer = photographer.Text
			next.ProfilePhoto = profilePhoto.Text
			next.Blog = strings.TrimSpace(website.Text)
			next.Website = strings.TrimSpace(website.Text)
			next.Github = github.Text
			err := next.Save()
			if err == nil {
				cfg = next
			}
			showSettingsSaveResult(status, "Author", err)
		},
	}
	return settingsPanel("About the photographer", "Profile and links displayed on your published website.", status, form,
		settingsGroup{"About page", form.Items[:3]},
		settingsGroup{"Social", form.Items[3:7]})
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
	return settingsPanel("Publishing", "Connect your Netlify site. Build and publish your gallery from Tasks.", status, form,
		settingsGroup{"Netlify", form.Items})
}

func uiConfigForm(server *preview.Server) fyne.CanvasObject {
	cfg := config.Config.UI
	status := newSettingsNotice()
	// Theme selection: Light or Dark
	themeOptions := []string{"light", "dark"}
	themeSelect := widget.NewRadioGroup(themeOptions, nil)
	themeSelect.Horizontal = true
	themeSelect.SetSelected(cfg.Theme)

	notifications := widget.NewCheck("Enable Notifications", nil)
	notifications.SetChecked(cfg.Notification)

	previewPublic := widget.NewCheck("Allow access from other devices", nil)
	previewPublic.SetChecked(cfg.Public)
	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.SetPlaceHolder("Enter Gemini API Key")
	apiKeyEntry.SetText(cfg.GeminiApiKey)
	aiEnabled := widget.NewCheck("Enable AI suggestions", func(enabled bool) {
		if enabled {
			apiKeyEntry.Enable()
		} else {
			apiKeyEntry.Disable()
		}
	})
	apiKeyEntry.Disable()
	aiEnabled.SetChecked(cfg.AIEnabled)

	form := &widget.Form{
		SubmitText: "Save",
		CancelText: "Reset",
		Items: []*widget.FormItem{
			{Text: "Theme", Widget: themeSelect, HintText: "Theme changes apply after restarting the application"},
			{Text: "Notifications", Widget: notifications, HintText: "Enable or disable notifications"},
			{Widget: previewPublic, HintText: "Share your website preview on your local network. When off, only this computer can access it."},
			{Widget: aiEnabled, HintText: "Optional and off by default. AI suggests titles, descriptions, and tags. It never edits your images or changes their appearance.\n\nWhen you request suggestions, the selected photo is sent to Google Gemini. Review the suggested text before saving."},
			{Text: "Gemini API Key", Widget: apiKeyEntry, HintText: "Used only when AI suggestions are enabled."},
		},
		OnCancel: func() {
			themeSelect.SetSelected(cfg.Theme)
			notifications.SetChecked(cfg.Notification)
			previewPublic.SetChecked(cfg.Public)
			apiKeyEntry.SetText(cfg.GeminiApiKey)
			aiEnabled.SetChecked(cfg.AIEnabled)
			status.Hide()
		},
		OnSubmit: func() {
			if themeSelect.Selected == "" {
				showSettingsSaveResult(status, "Application", fmt.Errorf("select a light or dark theme"))
				return
			}
			if aiEnabled.Checked && strings.TrimSpace(apiKeyEntry.Text) == "" {
				showSettingsSaveResult(status, "Application", fmt.Errorf("enter a Gemini API key or turn off AI suggestions"))
				return
			}

			next := cfg
			next.Theme = themeSelect.Selected
			next.Notification = notifications.Checked
			next.Public = previewPublic.Checked
			next.AIEnabled = aiEnabled.Checked
			next.GeminiApiKey = strings.TrimSpace(apiKeyEntry.Text)
			previewAccessChanged := next.Public != cfg.Public
			candidate := *config.Config
			candidate.UI = next
			err := candidate.Save()
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
			if err == nil && !next.AIEnabled {
				ai.ClearGeminiClient()
			} else if err == nil {
				if _, registerErr := ai.RegisterGeminiClient(); registerErr != nil {
					log.Printf("Could not configure Gemini: %v", registerErr)
				}
			}
		},
	}
	return settingsPanel("Application", "Appearance, browsing, notifications, and optional AI suggestions.", status, form,
		settingsGroup{"Appearance", form.Items[:2]},
		settingsGroup{"Preview", []*widget.FormItem{form.Items[2]}},
		settingsGroup{"AI", form.Items[3:]})
}
