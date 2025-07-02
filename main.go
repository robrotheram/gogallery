// Responsive grid layout for Fyne

package main

import (
	"embed"
	"testingFyne/cmd"
	"testingFyne/pkg/embeds"
)

//go:embed themes
var ThemeFS embed.FS

func init() {
	embeds.ThemeFS = ThemeFS
}

func main() {
	cmd.Execute()
}
