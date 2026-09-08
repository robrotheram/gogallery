package cmd

import (
	"fmt"
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"gogallery/pkg/monitor"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(benchmark)
}

var benchmark = &cobra.Command{
	Use: "benchmark",
	RunE: func(cmd *cobra.Command, args []string) error {
		cpuFile, err := os.Create("cpu.prof")
		if err != nil {
			return fmt.Errorf("create CPU profile: %w", err)
		}
		if err := pprof.StartCPUProfile(cpuFile); err != nil {
			_ = cpuFile.Close()
			return fmt.Errorf("start CPU profile: %w", err)
		}
		err = benchmarkScanPath()
		pprof.StopCPUProfile()
		if closeErr := cpuFile.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
		if err != nil {
			return err
		}

		memFile, err := os.Create("mem.prof")
		if err != nil {
			return fmt.Errorf("create memory profile: %w", err)
		}
		if err := pprof.WriteHeapProfile(memFile); err != nil {
			_ = memFile.Close()
			return fmt.Errorf("write memory profile: %w", err)
		}
		return memFile.Close()
	},
}

func benchmarkScanPath() error {

	start := time.Now()
	config, err := config.LoadConfig()
	if err != nil {
		return err
	}
	if err := config.Validate(); err != nil {
		return err
	}
	db, err := datastore.Open(datastore.DefaultDatabasePath, monitor.NewCMDMonitor())
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	if err := db.ScanPath(config.Gallery.Basepath); err != nil {
		return fmt.Errorf("scan gallery: %w", err)
	}

	elapsed := time.Since(start)
	log.Printf("Scan completed in %s", elapsed)
	return nil
}
