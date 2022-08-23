package cmd

import (
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
		datastore.Cache = &datastore.DataStore{}
		datastore.Cache.Open(config.Database.Baseurl)
		defer datastore.Cache.Close()

		worker.ScanPath(config.Gallery.Basepath)
		render := pipeline.NewRenderPipeline()
		// bar := progressbar.NewOptions(10,
		// 	progressbar.OptionEnableColorCodes(true),
		// 	progressbar.OptionShowBytes(false),
		// 	progressbar.OptionFullWidth(),
		// 	progressbar.OptionSetPredictTime(true),
		// 	progressbar.OptionSetElapsedTime(true),
		// 	progressbar.OptionOnCompletion(func() {
		// 		fmt.Printf("\n")
		// 	}))
		// bar.Describe("Building Site")
		render.BuildSite()
		//bar.Finish()
		return nil
	},
}
