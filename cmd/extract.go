package cmd

import (
	"fmt"
	"gogallery/pkg/embeds"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(templateCMD)
}

var templateCMD = &cobra.Command{
	Use:   "template",
	Short: "Extract template to directory",
	Long:  "Extract the internal template to any directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			fmt.Println("Please supply path to template directory")
			return nil
		}
		if err := embeds.CopyTheme(args[0]); err != nil {
			return err
		}
		fmt.Println("Theme extracted to: " + args[0])
		return nil
	},
}
