package cmd

import (
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/robrotheram/gogallery/backend/config"
	"github.com/robrotheram/gogallery/backend/datastore"
	"github.com/robrotheram/gogallery/backend/pipeline"
	templateengine "github.com/robrotheram/gogallery/backend/templateEngine"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(benchmark)
}

var benchmark = &cobra.Command{
	Use: "benchmark",
	RunE: func(cmd *cobra.Command, args []string) error {
		cpuFile, _ := os.Create("cpu.prof")
		pprof.StartCPUProfile(cpuFile)
		defer pprof.StopCPUProfile()

		memFile, _ := os.Create("mem.prof")
		pprof.WriteHeapProfile(memFile)
		defer memFile.Close()

		benchmarkScanPath()
		return nil
	},
}

func benchmarkScanPath() {

	start := time.Now()
	config := config.LoadConfig()
	config.Validate()
	db, err := datastore.Open(config.Gallery.Basepath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err := db.ScanPath(config.Gallery.Basepath); err != nil {
		log.Fatalf("Error scanning path: %v", err)
	}

	elapsed := time.Since(start)
	log.Printf("Scan completed in %s", elapsed)
}

func benchmarkImage() {
	var totalTime time.Duration
	p := datastore.Picture{
		Id:   "benchmark",
		Path: "/home/robert/Pictures/gallery/pictures/bergen/20250511_0010.jpg",
	}
	src, err := p.Load()
	if err != nil {
		log.Fatalf("Error loading benchmark image: %v", err)
	}
	destPath := "benchmark.webp"
	sizes := templateengine.ImageSizes
	for _, size := range sizes {
		if _, err := os.Stat(destPath); err == nil {
			if err := os.Remove(destPath); err != nil {
				log.Fatalf("Error deleting existing file: %v", err)
			}
		}
		fo, err := os.Create(destPath)
		if err != nil {
			log.Fatalf("Error creating file: %v", err)
		}
		defer fo.Close()

		start := time.Now()

		pipeline.ProcessImage(src, size.ImgWidth, fo)

		elapsed := time.Since(start)
		totalTime += elapsed
		log.Printf("Benchmark completed in %s", elapsed)
	}
	log.Printf("Total benchmark time for all sizes: %s", totalTime)
}
