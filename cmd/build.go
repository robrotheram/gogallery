package cmd

import (
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"gogallery/pkg/monitor"
	"gogallery/pkg/pipeline"
	"log"

	"github.com/gosuri/uiprogress"
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
		db, err := datastore.Open(config.Gallery.Basepath, cmdMonitor)
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		cmdMonitor.StartUpdater()
		db.ScanPath(config.Gallery.Basepath)
		log.Println("Building Site at: " + config.Gallery.Destpath)
		uiprogress.Start()
		render := pipeline.NewRenderPipeline(&config.Gallery, db)

		render.BuildSite()
		log.Println("Building Complete")
		return nil
	},
}
