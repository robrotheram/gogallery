package cmd

import (
	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
	"github.com/robrotheram/gogallery/serve"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serveCMD)
}

var serveCMD = &cobra.Command{
	Use:   "serve",
	Short: "serve static site",
	Long:  "serve static site",
	Run: func(cmd *cobra.Command, args []string) {

		config := config.LoadConfig()
		datastore.Cache = &datastore.DataStore{}
		datastore.Cache.Open(config.Gallery.Basepath)
		defer datastore.Cache.Close()
		if len(args) == 1 {
			config.Server.Port = ":" + args[0]
		}
		serve.Serve(config)
	},
}
