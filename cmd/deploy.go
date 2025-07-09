package cmd

import (
	"fmt"
	"gogallery/pkg/config"
	"gogallery/pkg/deploy"
	"gogallery/pkg/monitor"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deployCMD)
}

var deployCMD = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy static site",
	Long:  "Deploy static site",
	Run: func(cmd *cobra.Command, args []string) {
		config := config.LoadConfig()
		monitor := monitor.NewMonitor()
		config.Validate()
		fmt.Println("Deploying Site")
		deploy.DeploySite(*config, monitor.NewTask("deploy", 0))
	},
}
