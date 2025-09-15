package main

import (
	"data-comparator/internal/pkg/config"
	"data-comparator/internal/pkg/datareader"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <config.yaml>\n", os.Args[0])
		os.Exit(1)
	}
	
	configPath := os.Args[1]
	
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// Create data reader
	reader, err := datareader.New(cfg.Source)
	if err != nil {
		log.Fatalf("Failed to create data reader: %v", err)
	}
	defer reader.Close()
	
	fmt.Printf("Performance test started with config: %s\n", configPath)
	fmt.Printf("Press Ctrl+C to stop...\n\n")
	
	start := time.Now()
	recordCount := int64(0)
	lastReport := start
	reportInterval := 5 * time.Second
	
	// Monitor memory usage
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)
	
	// Read records and track performance
	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Reached end of stream after %d records.\n", recordCount)
				break
			}
			log.Fatalf("Error reading record: %v", err)
		}
		
		if record == nil {
			continue
		}
		
		recordCount++
		
		// Report progress every few seconds
		now := time.Now()
		if now.Sub(lastReport) >= reportInterval {
			elapsed := now.Sub(start)
			rate := float64(recordCount) / elapsed.Seconds()
			
			runtime.ReadMemStats(&m2)
			memUsedMB := float64(m2.Alloc-m1.Alloc) / 1024 / 1024
			
			fmt.Printf("Records: %d, Rate: %.0f/s, Elapsed: %v, Memory: +%.1f MB\n", 
				recordCount, rate, elapsed.Round(time.Second), memUsedMB)
			
			lastReport = now
		}
	}
	
	// Final statistics
	totalElapsed := time.Since(start)
	finalRate := float64(recordCount) / totalElapsed.Seconds()
	
	runtime.ReadMemStats(&m2)
	finalMemUsedMB := float64(m2.Alloc-m1.Alloc) / 1024 / 1024
	
	fmt.Printf("\n=== Final Statistics ===\n")
	fmt.Printf("Total Records: %d\n", recordCount)
	fmt.Printf("Total Time: %v\n", totalElapsed.Round(time.Millisecond))
	fmt.Printf("Average Rate: %.2f records/second\n", finalRate)
	fmt.Printf("Memory Usage: +%.1f MB\n", finalMemUsedMB)
}