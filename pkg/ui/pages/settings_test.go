package pages

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/spf13/viper"
	"gogallery/pkg/ai"
	"gogallery/pkg/config"
)

func TestShowSettingsSaveResultReportsSuccessInline(t *testing.T) {
	app := test.NewApp()
	t.Cleanup(app.Quit)
	previousNotifications := config.Config.UI.Notification
	config.Config.UI.Notification = false
	t.Cleanup(func() { config.Config.UI.Notification = previousNotifications })

	notice := newSettingsNotice()
	t.Cleanup(notice.Hide)
	showSettingsSaveResult(notice, "Gallery", nil)

	if notice.title.Text != "Changes saved" {
		t.Fatalf("unexpected save title: %q", notice.title.Text)
	}
	if notice.message.Text != "Gallery settings are up to date." {
		t.Fatalf("unexpected save message: %q", notice.message.Text)
	}
	if notice.title.Importance != widget.MediumImportance {
		t.Fatalf("expected neutral importance, got %v", notice.title.Importance)
	}
	if !notice.container.Visible() {
		t.Fatal("expected success notice to be visible")
	}
	notice.timerMu.Lock()
	timerScheduled := notice.dismissTimer != nil
	notice.timerMu.Unlock()
	if !timerScheduled {
		t.Fatal("expected success notice to schedule automatic dismissal")
	}
}

func TestShowSettingsSaveResultReportsFailureInline(t *testing.T) {
	app := test.NewApp()
	t.Cleanup(app.Quit)
	previousNotifications := config.Config.UI.Notification
	config.Config.UI.Notification = false
	t.Cleanup(func() { config.Config.UI.Notification = previousNotifications })

	notice := newSettingsNotice()
	t.Cleanup(notice.Hide)
	showSettingsSaveResult(notice, "Gallery", errors.New("disk is read-only"))

	if notice.title.Text != "Could not save settings" {
		t.Fatalf("unexpected failure title: %q", notice.title.Text)
	}
	if notice.message.Text != "disk is read-only" {
		t.Fatalf("unexpected failure message: %q", notice.message.Text)
	}
	if notice.title.Importance != widget.DangerImportance {
		t.Fatalf("expected danger importance, got %v", notice.title.Importance)
	}
	if !notice.container.Visible() {
		t.Fatal("expected failure notice to be visible")
	}
}

func TestHideSettingsNotice(t *testing.T) {
	app := test.NewApp()
	t.Cleanup(app.Quit)
	notice := newSettingsNotice()
	t.Cleanup(notice.Hide)
	showSettingsSaveResult(notice, "Gallery", nil)

	notice.Hide()

	if notice.container.Visible() {
		t.Fatal("expected notice to be hidden")
	}
	notice.timerMu.Lock()
	timerScheduled := notice.dismissTimer != nil
	notice.timerMu.Unlock()
	if timerScheduled {
		t.Fatal("expected manual dismissal to cancel automatic dismissal")
	}
}

func TestSettingsNoticeDismissesAfterTwentySeconds(t *testing.T) {
	if settingsNoticeDuration != 20*time.Second {
		t.Fatalf("expected a 20 second dismissal delay, got %v", settingsNoticeDuration)
	}
}

func TestAISettingsRequireSavingExplicitOptIn(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	viper.Reset()
	t.Cleanup(viper.Reset)
	previous := config.Config
	config.Config = &config.Configuration{UI: config.UIConfiguration{Theme: "dark", ImagesPerPage: 20, GeminiApiKey: "saved-test-key"}}
	t.Cleanup(func() { config.Config = previous; ai.ClearGeminiClient() })
	path := filepath.Join(t.TempDir(), "settings.yaml")
	if err := os.WriteFile(path, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	viper.SetConfigFile(path)
	panel := uiConfigForm(nil)
	window := test.NewWindow(panel)
	defer window.Close()
	var enabled *widget.Check
	var key *widget.Entry
	var save, reset *widget.Button
	var visit func(fyne.CanvasObject)
	visit = func(object fyne.CanvasObject) {
		switch o := object.(type) {
		case *widget.Label:
			if o.Text == "Images Per Page" {
				t.Error("Application settings must not expose a desktop images-per-page limit")
			}
		case *widget.Check:
			if o.Text == "Enable AI suggestions" {
				enabled = o
			}
		case *widget.Entry:
			if o.PlaceHolder == "Enter Gemini API Key" {
				key = o
			}
		case *widget.Button:
			if o.Text == "Save changes" {
				save = o
			} else if o.Text == "Reset changes" {
				reset = o
			}
		}
		if c, ok := object.(*fyne.Container); ok {
			for _, child := range c.Objects {
				visit(child)
			}
		} else if w, ok := object.(fyne.Widget); ok {
			for _, child := range test.WidgetRenderer(w).Objects() {
				visit(child)
			}
		}
	}
	visit(panel)
	if enabled == nil || key == nil || save == nil || reset == nil {
		t.Fatal("AI preference controls were not found")
	}
	defer reset.OnTapped() // also stops the saved/error notice timer
	if enabled.Checked || !key.Disabled() || ai.IsAIEnabled() {
		t.Fatal("AI must remain off despite a saved API key")
	}
	enabled.SetChecked(true)
	if key.Disabled() || config.Config.UI.AIEnabled || ai.IsAIEnabled() {
		t.Fatal("unsaved opt-in should enable only the key field")
	}
	reset.OnTapped()
	if enabled.Checked || !key.Disabled() {
		t.Fatal("Reset did not restore the saved disabled preference")
	}
	enabled.SetChecked(true)
	key.SetText("")
	save.OnTapped()
	if config.Config.UI.AIEnabled || ai.IsAIEnabled() {
		t.Fatal("AI was enabled without an API key")
	}
	key.SetText("saved-test-key")
	save.OnTapped()
	if !config.Config.UI.AIEnabled || !ai.IsAIEnabled() {
		t.Fatal("saving explicit opt-in did not enable suggestions")
	}
	enabled.SetChecked(false)
	save.OnTapped()
	if config.Config.UI.AIEnabled || ai.IsAIEnabled() || !key.Disabled() || config.Config.UI.GeminiApiKey != "saved-test-key" {
		t.Fatal("turning AI off must stop suggestions while preserving the saved key")
	}
}

func TestAuthorSettingsCombineWebsiteAndBlog(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	viper.Reset()
	t.Cleanup(viper.Reset)
	previous := config.Config
	config.Config = &config.Configuration{About: config.AboutConfiguration{Blog: "https://old.example", Facebook: "https://facebook.com/saved", Email: "saved@example.com"}}
	t.Cleanup(func() { config.Config = previous })
	path := filepath.Join(t.TempDir(), "settings.yaml")
	if err := os.WriteFile(path, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	viper.SetConfigFile(path)
	panel := aboutConfigForm()
	var website *widget.Entry
	var save, reset *widget.Button
	labels := map[string]bool{}
	var visit func(fyne.CanvasObject)
	visit = func(object fyne.CanvasObject) {
		switch o := object.(type) {
		case *widget.Label:
			labels[o.Text] = true
		case *widget.Entry:
			if o.Text == "https://old.example" {
				website = o
			}
			if o.Text == "saved@example.com" || o.Text == "https://facebook.com/saved" {
				t.Error("removed contact field is still displayed")
			}
		case *widget.Button:
			if o.Text == "Save changes" {
				save = o
			}
			if o.Text == "Reset changes" {
				reset = o
			}
		}
		if c, ok := object.(*fyne.Container); ok {
			for _, child := range c.Objects {
				visit(child)
			}
		} else if w, ok := object.(fyne.Widget); ok {
			for _, child := range test.WidgetRenderer(w).Objects() {
				visit(child)
			}
		}
	}
	visit(panel)
	if website == nil || save == nil || reset == nil || !labels["Website"] || !labels["X.com"] {
		t.Fatal("missing combined website or X.com controls")
	}
	for _, removed := range []string{"Footer", "Background Photo", "Website / Blog", "Facebook", "Email"} {
		if labels[removed] {
			t.Errorf("removed field %q is still displayed", removed)
		}
	}
	defer reset.OnTapped()
	website.SetText("https://new.example")
	reset.OnTapped()
	if website.Text != "https://old.example" {
		t.Fatal("reset lost legacy blog URL")
	}
	website.SetText("https://new.example")
	save.OnTapped()
	if config.Config.About.Website != "https://new.example" || config.Config.About.Blog != "https://new.example" {
		t.Fatal("combined link was not saved")
	}
	if config.Config.About.Email != "saved@example.com" || config.Config.About.Facebook != "https://facebook.com/saved" {
		t.Fatal("hidden legacy contact data was changed")
	}
}
