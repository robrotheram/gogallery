package ui

import (
	"image/color"
	"math"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestDarkThemeForegroundIsWhite(t *testing.T) {
	foreground := NewComfortableTheme("dark").Color(theme.ColorNameForeground, theme.VariantDark)
	if foreground != color.White {
		t.Fatalf("dark foreground = %v, want white", foreground)
	}
}

func TestTaskStatusColoursHaveReadableContrast(t *testing.T) {
	for _, variant := range []string{"light", "dark"} {
		t.Run(variant, func(t *testing.T) {
			palette := NewComfortableTheme(variant)
			for _, foreground := range []fyne.ThemeColorName{theme.ColorNameSuccess, theme.ColorNameError, theme.ColorNameForeground} {
				for _, background := range []fyne.ThemeColorName{theme.ColorNameBackground, theme.ColorNameButton} {
					ratio := contrastRatio(palette.Color(foreground, theme.VariantLight), palette.Color(background, theme.VariantLight))
					if ratio < 4.5 {
						t.Errorf("%s on %s: contrast %.2f is below 4.5", foreground, background, ratio)
					}
				}
			}
			for _, pair := range [][2]fyne.ThemeColorName{
				{theme.ColorNameForegroundOnSuccess, theme.ColorNameSuccess},
				{theme.ColorNameForegroundOnError, theme.ColorNameError},
			} {
				if ratio := contrastRatio(palette.Color(pair[0], theme.VariantLight), palette.Color(pair[1], theme.VariantLight)); ratio < 4.5 {
					t.Errorf("%s on %s: button contrast %.2f is below 4.5", pair[0], pair[1], ratio)
				}
			}
		})
	}
}

func contrastRatio(a, b color.Color) float64 {
	luminance := func(c color.Color) float64 {
		r, g, b, _ := c.RGBA()
		linear := func(channel uint32) float64 {
			s := float64(channel) / 65535
			if s <= 0.04045 {
				return s / 12.92
			}
			return math.Pow((s+0.055)/1.055, 2.4)
		}
		return 0.2126*linear(r) + 0.7152*linear(g) + 0.0722*linear(b)
	}
	x, y := luminance(a), luminance(b)
	return (max(x, y) + 0.05) / (min(x, y) + 0.05)
}
