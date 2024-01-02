package cmd

import (
	"github.com/robrotheram/gogallery/backend/api"
	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/datastore"
	"github.com/robrotheram/gogallery/backend/embeds"
	"github.com/spf13/cobra"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
)

func init() {
	rootCmd.AddCommand(dashboadCMD)
}

var dashboadCMD = &cobra.Command{
	Use:   "dashboard",
	Short: "Launch UI",
	Long:  "Launch UI",
	Run: func(cmd *cobra.Command, args []string) {
		LaunchDashboard()
	},
}

func LaunchDashboard() error {
	config := config.LoadConfig()
	db := datastore.Open(config.Gallery.Basepath)
	defer db.Close()
	config.Server.Port = "8800"
	go api.NewGoGalleryAPI(config, db).DashboardAPI()

	return wails.Run(&options.App{
		Title:            "gogallery",
		Width:            1024,
		Height:           768,
		MinWidth:         1200,
		Assets:           &embeds.DashboardFS,
		Frameless:        false,
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
	})
}
