package cmd

import (
	"fmt"
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"gogallery/pkg/monitor"
	"gogallery/pkg/preview"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serveCMD)
}

var serveCMD = &cobra.Command{
	Use:   "serve",
	Short: "Serve static site",
	Long:  "Serve static site",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}
		if err := cfg.Validate(); err != nil {
			return err
		}
		db, err := datastore.Open(datastore.DefaultDatabasePath, monitor.NewCMDMonitor())
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()

		server := preview.NewServer(db)
		if err := server.Start(); err != nil {
			return fmt.Errorf("start server: %w", err)
		}
		// Print the actual address after the server has started and acquired a port
		log.Printf("Starting Preview Server http://%s", server.Addr())
		ctx, stopSignals := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stopSignals()
		<-ctx.Done()
		if err := server.Stop(); err != nil {
			return fmt.Errorf("stop server: %w", err)
		}
		return nil
	},
}
