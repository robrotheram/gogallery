package cmd

import (
	"fmt"

	"github.com/robrotheram/gogallery/embeds"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(templateCMD)
}

var templateCMD = &cobra.Command{
	Use:   "template",
	Short: "extract template",
	Long:  "extract the internal template to any directory",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Please suply path to template")
			return
		}
		embeds.CopyTheme(args[0])
		fmt.Println("Theme extracted to: " + args[0])
	},
}
