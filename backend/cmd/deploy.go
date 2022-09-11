package cmd

import (
	"fmt"

	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/deploy"
	"github.com/robrotheram/gogallery/backend/pipeline"
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
		monitor := pipeline.NewMonitor()
		config.Validate()
		fmt.Println("Deploying Site")
		deploy.DeploySite(*config, monitor.NewTask("deploy"))
	},
}
