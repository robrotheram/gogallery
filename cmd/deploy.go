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
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.LoadConfig()
		if err != nil {
			return err
		}
		monitor := monitor.NewMonitor()
		if err := config.Validate(); err != nil {
			return err
		}
		fmt.Println("Deploying Site")
		return deploy.DeploySite(*config, monitor.NewTask("deploy", 0))
	},
}
