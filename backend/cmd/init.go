package cmd

import (
	"github.com/k0kubun/pp/v3"
	"github.com/manifoldco/promptui"
	"github.com/robrotheram/gogallery/backend/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCMD)
}

var initCMD = &cobra.Command{
	Use:   "init",
	Short: "create site",
	Long:  "create site",
	Run: func(cmd *cobra.Command, args []string) {
		config := config.LoadConfig()
		config.PromptSiteName()
		config.PromptGalleryBasePath()
		config.PromptGalleryDest()
		config.PromptGalleryTheme()

		pp.Print(config)
		prompt := promptui.Prompt{
			Label:     "Is info correct",
			IsConfirm: true,
		}
		prompt.Run()
		config.Save()
	},
}
