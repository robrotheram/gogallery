package cmd

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"

	"github.com/robrotheram/gogallery/backend/api"
	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/datastore"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serveCMD)
	rootCmd.AddCommand(devCMD)
}

var serveCMD = &cobra.Command{
	Use:   "serve",
	Short: "Serve static site",
	Long:  "Serve static site",
	Run: func(cmd *cobra.Command, args []string) {

		config := config.LoadConfig()
		db := datastore.Open(config.Gallery.Basepath)
		defer db.Close()
		db.ScanPath(config.Gallery.Basepath)
		if len(args) == 1 {
			config.Server.Port = ":" + args[0]
		}
		openbrowser(fmt.Sprintf("http://%s", config.Server.GetLocalAddr()))
		api.NewGoGalleryAPI(config, db).Serve()
	},
}

var devCMD = &cobra.Command{
	Use: "dev",
	Run: func(cmd *cobra.Command, args []string) {
		config := config.LoadConfig()
		db := datastore.Open(config.Gallery.Basepath)
		defer db.Close()
		db.ScanPath(config.Gallery.Basepath)
		config.Server.Port = "8800"
		api.NewGoGalleryAPI(config, db).DashboardAPI()
	},
}

func openbrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}
