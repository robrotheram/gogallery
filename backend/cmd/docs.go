package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func init() {
	rootCmd.AddCommand(docCMD)
}

var docCMD = &cobra.Command{
	Use:   "docs",
	Short: "cli docs",
	Long:  "Update cli documentation",
	Run: func(cmd *cobra.Command, args []string) {
		os.MkdirAll("./docs/cli", os.ModePerm)
		err := doc.GenMarkdownTree(rootCmd, "./docs/cli")
		if err != nil {
			fmt.Println(err)
		}
	},
}
