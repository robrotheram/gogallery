package cmd

import (
	"fmt"
	"testingFyne/pkg/embeds"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(templateCMD)
}

var templateCMD = &cobra.Command{
	Use:   "template",
	Short: "Extract template to directory",
	Long:  "Extract the internal template to any directory",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Please supply path to template directory")
			return
		}
		embeds.CopyTheme(args[0])
		fmt.Println("Theme extracted to: " + args[0])
	},
}
