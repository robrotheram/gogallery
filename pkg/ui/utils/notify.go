package utils

import (
	"gogallery/pkg/config"

	"fyne.io/fyne/v2"
)

func Notify(title string, message string) {
	if config.Config.UI.Notification {
		app := fyne.CurrentApp()
		if app == nil {
			return
		}
		app.SendNotification(&fyne.Notification{
			Title:   title,
			Content: message,
		})
	}
}
