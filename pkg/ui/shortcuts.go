package ui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

func installCloseShortcut(window fyne.Window, quit func()) {
	shortcut := newCloseShortcut()
	requestQuit := func() {
		log.Println("Close shortcut pressed; shutting down")
		quit()
	}
	window.Canvas().AddShortcut(shortcut, func(fyne.Shortcut) {
		requestQuit()
	})
}

func newCloseShortcut() *desktop.CustomShortcut {
	return &desktop.CustomShortcut{
		KeyName:  fyne.KeyW,
		Modifier: fyne.KeyModifierShortcutDefault,
	}
}
