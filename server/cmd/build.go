package cmd

import (
	"log"

	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
	"github.com/robrotheram/gogallery/pipeline"
	"github.com/robrotheram/gogallery/worker"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildCMD)
}

var buildCMD = &cobra.Command{
	Use:   "build",
	Short: "build static site",
	Long:  "build static site",
	RunE: func(cmd *cobra.Command, args []string) error {

		config := config.LoadConfig()
		config.Validate()
		datastore.Cache = &datastore.DataStore{}
		datastore.Cache.Open(config.Gallery.Basepath)
		defer datastore.Cache.Close()

		worker.ScanPath(config.Gallery.Basepath)
		log.Println("Building Site at: " + config.Gallery.Destpath)
		render := pipeline.NewRenderPipeline(config.Gallery.Destpath)
		render.BuildSite()
		log.Println("Building Complete")
		return nil
	},
}
