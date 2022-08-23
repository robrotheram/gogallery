package cmd

import (
	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(adminCMD)
}

var adminCMD = &cobra.Command{
	Use:   "reset-password",
	Short: "reset admin password",
	Long:  "reset admin password",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := config.LoadConfig()
		datastore.Cache = &datastore.DataStore{}
		datastore.Cache.Open(config.Gallery.Basepath)
		defer datastore.Cache.Close()
		datastore.CreateDefaultUser()
		return nil
	},
}
