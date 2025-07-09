package utils

import (
	"gogallery/pkg/config"

	"fyne.io/fyne/v2"
)

func Notify(tite string, msg string) {
	if config.Config.UI.Notification {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   tite,
			Content: msg,
		})
	}
}
