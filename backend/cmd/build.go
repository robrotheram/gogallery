package cmd

import (
	"log"

	"github.com/gosuri/uiprogress"
	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/datastore"
	"github.com/robrotheram/gogallery/backend/monitor"
	"github.com/robrotheram/gogallery/backend/pipeline"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildCMD)
}

var cmdMonitor = monitor.NewCMDMonitor()
var buildCMD = &cobra.Command{
	Use:   "build",
	Short: "Build static site",
	Long:  "Build static site",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := config.LoadConfig()
		config.Validate()
		db, err := datastore.Open(config.Gallery.Basepath)
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		db.ScanPath(config.Gallery.Basepath)
		log.Println("Building Site at: " + config.Gallery.Destpath)
		uiprogress.Start()
		render := pipeline.NewRenderPipeline(&config.Gallery, db, cmdMonitor)
		cmdMonitor.StartUpdater()
		render.BuildSite()
		cmdMonitor.StopUpdater()
		log.Println("Building Complete")
		return nil
	},
}
