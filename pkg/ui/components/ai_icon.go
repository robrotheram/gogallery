package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// AIIcon uses sparkles to distinguish AI suggestions from adding content.
func AIIcon() fyne.Resource {
	return theme.NewThemedResource(aiIconResource)
}

var aiIconResource = fyne.NewStaticResource("ai-sparkles.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#000" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="m12 3 2.5 6.5L21 12l-6.5 2.5L12 21l-2.5-6.5L3 12l6.5-2.5Z"/><path d="M20 2v4m-2-2h4M4 18v4m-2-2h4"/></svg>`))
