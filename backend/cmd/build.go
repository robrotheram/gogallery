package cmd

import (
	"log"

	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/datastore"
	"github.com/robrotheram/gogallery/backend/pipeline"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildCMD)
}

var buildCMD = &cobra.Command{
	Use:   "build",
	Short: "Build static site",
	Long:  "Build static site",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := config.LoadConfig()
		config.Validate()

		// templateengine.TemplateBuilder()

		db := datastore.Open(config.Gallery.Basepath)
		defer db.Close()
		db.ScanPath(config.Gallery.Basepath)
		log.Println("Building Site at: " + config.Gallery.Destpath)
		monitor := pipeline.NewMonitor()
		render := pipeline.NewRenderPipeline(&config.Gallery, db, monitor)
		render.BuildSite()
		log.Println("Building Complete")
		return nil
	},
}
