// package cmd

// import (
// 	"github.com/robrotheram/gogallery/backend/config"
// 	"github.com/robrotheram/gogallery/backend/datastore"
// 	"github.com/spf13/cobra"
// )

// func init() {
// 	rootCmd.AddCommand(adminCMD)
// }

// var adminCMD = &cobra.Command{
// 	Use:   "reset-password",
// 	Short: "reset admin password",
// 	Long:  "reset admin password",
// 	RunE: func(cmd *cobra.Command, args []string) error {
// 		config := config.LoadConfig()
// 		db := datastore.Open(config.Gallery.Basepath)
// 		defer db.Close()
// 		datastore.CreateDefaultUser()
// 		return nil
// 	},
// }

models