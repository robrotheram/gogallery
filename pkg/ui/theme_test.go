package ui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2/theme"
)

func TestDarkThemeForegroundIsWhite(t *testing.T) {
	foreground := NewComfortableTheme("dark").Color(theme.ColorNameForeground, theme.VariantDark)
	if foreground != color.White {
		t.Fatalf("dark foreground = %v, want white", foreground)
	}
}
