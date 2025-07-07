package cmd

import (
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"gogallery/pkg/monitor"
	"gogallery/pkg/preview"
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serveCMD)
}

var serveCMD = &cobra.Command{
	Use:   "serve",
	Short: "Serve static site",
	Long:  "Serve static site",
	Run: func(cmd *cobra.Command, args []string) {
		config := config.LoadConfig()
		db, err := datastore.Open(config.Gallery.Basepath, monitor.NewCMDMonitor())
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}

		server := preview.NewServer(db)
		if err := server.Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
		// Print the actual address after the server has started and acquired a port
		log.Printf("Starting Preview Server http://%s", server.Addr())
		// Wait for the server goroutine to exit (block until server stops)
		select {}
	},
}
