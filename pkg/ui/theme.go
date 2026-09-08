package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ComfortableTheme is a custom theme with increased spacing and padding for a less compact UI
type ComfortableTheme struct {
	fyne.Theme
	variant string // "light" or "dark"
}

var darkPalette = map[fyne.ThemeColorName]color.Color{
	theme.ColorNameButton:              color.NRGBA{R: 35, G: 41, B: 45, A: 255},
	theme.ColorNameDisabledButton:      color.NRGBA{R: 28, G: 33, B: 36, A: 255},
	theme.ColorNamePrimary:             color.NRGBA{R: 45, G: 178, B: 111, A: 255},
	theme.ColorNameForegroundOnPrimary: color.White,
	theme.ColorNameSelection:           color.NRGBA{R: 45, G: 178, B: 111, A: 110},
	theme.ColorNameFocus:               color.NRGBA{R: 79, G: 209, B: 143, A: 180},
	theme.ColorNameBackground:          color.NRGBA{R: 17, G: 20, B: 22, A: 255},
	theme.ColorNameForeground:          color.White,
	theme.ColorNameInputBackground:     color.NRGBA{R: 27, G: 32, B: 35, A: 255},
	theme.ColorNameInputBorder:         color.NRGBA{R: 69, G: 79, B: 84, A: 255},
	theme.ColorNameScrollBar:           color.NRGBA{R: 103, G: 116, B: 122, A: 210},
	theme.ColorNameShadow:              color.NRGBA{R: 0, G: 0, B: 0, A: 100},
	theme.ColorNameSeparator:           color.NRGBA{R: 62, G: 70, B: 74, A: 180},
	theme.ColorNameMenuBackground:      color.NRGBA{R: 21, G: 25, B: 27, A: 255},
	theme.ColorNameOverlayBackground:   color.NRGBA{R: 9, G: 11, B: 12, A: 220},
	theme.ColorNameHover:               color.NRGBA{R: 255, G: 255, B: 255, A: 18},
}

var lightPalette = map[fyne.ThemeColorName]color.Color{
	theme.ColorNameButton:              color.NRGBA{R: 231, G: 237, B: 240, A: 255},
	theme.ColorNameDisabledButton:      color.NRGBA{R: 240, G: 243, B: 244, A: 255},
	theme.ColorNamePrimary:             color.NRGBA{R: 22, G: 142, B: 86, A: 255},
	theme.ColorNameForegroundOnPrimary: color.White,
	theme.ColorNameSelection:           color.NRGBA{R: 22, G: 142, B: 86, A: 70},
	theme.ColorNameFocus:               color.NRGBA{R: 22, G: 142, B: 86, A: 140},
	theme.ColorNameBackground:          color.NRGBA{R: 247, G: 249, B: 250, A: 255},
	theme.ColorNameForeground:          color.NRGBA{R: 27, G: 35, B: 39, A: 255},
	theme.ColorNameInputBackground:     color.White,
	theme.ColorNameInputBorder:         color.NRGBA{R: 190, G: 200, B: 205, A: 255},
	theme.ColorNameScrollBar:           color.NRGBA{R: 102, G: 115, B: 121, A: 180},
	theme.ColorNameShadow:              color.NRGBA{R: 23, G: 32, B: 36, A: 45},
	theme.ColorNameSeparator:           color.NRGBA{R: 205, G: 213, B: 217, A: 255},
	theme.ColorNameMenuBackground:      color.NRGBA{R: 241, G: 245, B: 246, A: 255},
	theme.ColorNameOverlayBackground:   color.NRGBA{R: 245, G: 248, B: 249, A: 235},
	theme.ColorNameHover:               color.NRGBA{R: 18, G: 31, B: 37, A: 18},
}

// NewComfortableTheme creates a new ComfortableTheme with the specified variant
func NewComfortableTheme(variant string) *ComfortableTheme {
	return &ComfortableTheme{
		variant: variant,
	}
}

func (c *ComfortableTheme) Color(name fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	palette := lightPalette
	if c.variant == "dark" {
		palette = darkPalette
	}
	if col, ok := palette[name]; ok {
		return col
	}
	return theme.DefaultTheme().Color(name, v)
}

func (c *ComfortableTheme) Font(style fyne.TextStyle) fyne.Resource {
	// Use default theme fonts
	return theme.DefaultTheme().Font(style)
}

func (c *ComfortableTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	// Use default theme icons
	return theme.DefaultTheme().Icon(name)
}

func (c *ComfortableTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInnerPadding:
		return 8
	case theme.SizeNameInputRadius:
		return 8
	case theme.SizeNameSelectionRadius:
		return 8
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameText:
		return 15
	case theme.SizeNameScrollBar:
		return 8
	case theme.SizeNameScrollBarSmall:
		return 4
	default:
		return theme.DefaultTheme().Size(name)
	}
}
func (c *ComfortableTheme) Variant() string {
	return c.variant
}
