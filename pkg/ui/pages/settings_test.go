package pages

import (
	"errors"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
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
