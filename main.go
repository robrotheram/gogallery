package main

import (
	"embed"

	"github.com/robrotheram/gogallery/backend/cmd"
	"github.com/robrotheram/gogallery/backend/embeds"
)

//go:embed frontend/dist
var assets embed.FS

//go:embed themes/eastnor
var ThemeFS embed.FS

func init() {
	embeds.ThemeFS = ThemeFS
	embeds.DashboardFS = assets
}

func main() {
	cmd.Execute()
}
