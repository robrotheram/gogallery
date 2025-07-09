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

// NewComfortableTheme creates a new ComfortableTheme with the specified variant
func NewComfortableTheme(variant string) *ComfortableTheme {
	return &ComfortableTheme{
		variant: variant,
	}
}

func (c *ComfortableTheme) Color(name fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	// Refactored: use a map for color lookups to reduce complexity
	isDark := c.variant == "dark"
	if isDark {
		colorMap := map[fyne.ThemeColorName]color.Color{
			theme.ColorNameButton:            color.RGBA{R: 0, G: 168, B: 89, A: 255},    // Rich Green (dark)
			theme.ColorNamePrimary:           color.RGBA{R: 0, G: 168, B: 89, A: 255},    // Less bright green for progress bar
			theme.ColorNameSelection:         color.RGBA{R: 0, G: 168, B: 89, A: 180},    // Green selection for list highlight (dark)
			theme.ColorNameBackground:        color.RGBA{R: 8, G: 12, B: 8, A: 255},      // Nearly black background (dark)
			theme.ColorNameForeground:        color.RGBA{R: 255, G: 255, B: 255, A: 255}, // White text
			theme.ColorNameInputBackground:   color.RGBA{R: 20, G: 28, B: 24, A: 255},    // Deep, neutral green-gray for input background
			theme.ColorNameInputBorder:       color.RGBA{R: 0, G: 100, B: 50, A: 255},    // Darker green border
			theme.ColorNameScrollBar:         color.RGBA{R: 0, G: 168, B: 89, A: 255},    // Consistent green scroll bar
			theme.ColorNameShadow:            color.RGBA{R: 0, G: 0, B: 0, A: 0},         // Transparent shadows
			theme.ColorNameSeparator:         color.RGBA{R: 30, G: 30, B: 30, A: 200},    // Blackish separator
			theme.ColorNameMenuBackground:    color.RGBA{R: 12, G: 24, B: 12, A: 255},    // Sidebar bg
			theme.ColorNameOverlayBackground: color.RGBA{R: 0, G: 200, B: 83, A: 100},    // Card border
			theme.ColorNameHover:             color.RGBA{R: 40, G: 40, B: 40, A: 10},     // Darker gray for button/input hover
		}
		if col, ok := colorMap[name]; ok {
			return col
		}
	} else {
		colorMapLight := map[fyne.ThemeColorName]color.Color{
			theme.ColorNameButton:            color.RGBA{R: 0, G: 168, B: 89, A: 255},    // Rich Green (light)
			theme.ColorNamePrimary:           color.RGBA{R: 0, G: 168, B: 89, A: 255},    // Less bright green for progress bar
			theme.ColorNameSelection:         color.RGBA{R: 0, G: 168, B: 89, A: 180},    // Green selection for list highlight (light)
			theme.ColorNameBackground:        color.RGBA{R: 250, G: 252, B: 245, A: 255}, // Offwhite background (light)
			theme.ColorNameForeground:        color.RGBA{R: 30, G: 60, B: 30, A: 255},    // Deep green text
			theme.ColorNameInputBackground:   color.RGBA{R: 240, G: 255, B: 240, A: 255}, // Slightly green-tinted offwhite
			theme.ColorNameInputBorder:       color.RGBA{R: 0, G: 200, B: 83, A: 255},    // Consistent Rich Green border
			theme.ColorNameScrollBar:         color.RGBA{R: 0, G: 168, B: 89, A: 255},    // Consistent green scroll bar
			theme.ColorNameShadow:            color.RGBA{R: 0, G: 0, B: 0, A: 0},         // Transparent shadows
			theme.ColorNameSeparator:         color.RGBA{R: 30, G: 30, B: 30, A: 200},    // Blackish separator
			theme.ColorNameMenuBackground:    color.RGBA{R: 220, G: 255, B: 220, A: 255}, // Sidebar bg
			theme.ColorNameOverlayBackground: color.RGBA{R: 0, G: 200, B: 83, A: 100},    // Card border
			theme.ColorNameHover:             color.RGBA{R: 40, G: 40, B: 40, A: 100},    // Darker gray for button/input hover
		}
		if col, ok := colorMapLight[name]; ok {
			return col
		}
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
		return 10 // Increased padding for a more comfortable layout
	case theme.SizeNameInnerPadding:
		return 12 // Increased inner padding for select and similar widgets
	case theme.SizeNameInputRadius:
		return 14
	case theme.SizeNameSeparatorThickness:
		return 1 // Thicker separators for sidebar borders
	case theme.SizeNameText:
		return 16 // Slightly larger text for better readability
	case theme.SizeNameScrollBar:
		return 4 // Thicker scroll bar for easier interaction
	case theme.SizeNameScrollBarSmall:
		return 2 // Smaller scroll bar
	default:
		return theme.DefaultTheme().Size(name)
	}
}
func (c *ComfortableTheme) Variant() string {
	return c.variant
}
