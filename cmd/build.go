package cmd

import (
	"fmt"
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
		config, err := config.LoadConfig()
		if err != nil {
			return err
		}
		if err := config.Validate(); err != nil {
			return err
		}
		db, err := datastore.Open(datastore.DefaultDatabasePath, cmdMonitor)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()
		cmdMonitor.StartUpdater()
		if err := db.ScanPath(config.Gallery.Basepath); err != nil {
			return err
		}
		log.Println("Building Site at: " + config.Gallery.Destpath)
		uiprogress.Start()
		render := pipeline.NewRenderPipeline(&config.Gallery, db)

		if err := render.BuildSite(); err != nil {
			return err
		}
		log.Println("Building Complete")
		return nil
	},
}
