package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestInstallCloseShortcut(t *testing.T) {
	app := test.NewApp()
	window := app.NewWindow("GoGallery")
	quitCalled := false

	installCloseShortcut(window, func() {
		quitCalled = true
	})

	if mainMenu := window.MainMenu(); mainMenu != nil {
		t.Fatalf("close shortcut should not create a visible main menu: %#v", mainMenu)
	}

	shortcut := newCloseShortcut()
	if shortcut.Key() != fyne.KeyW {
		t.Errorf("shortcut key = %q, want %q", shortcut.Key(), fyne.KeyW)
	}
	if shortcut.Mod() != fyne.KeyModifierShortcutDefault {
		t.Errorf("shortcut modifier = %v, want %v", shortcut.Mod(), fyne.KeyModifierShortcutDefault)
	}
	if quitCalled {
		t.Fatal("installing the shortcut should not immediately quit")
	}
}
