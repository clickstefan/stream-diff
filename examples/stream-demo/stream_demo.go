package main

import (
	"data-comparator/internal/pkg/config"
	"data-comparator/internal/pkg/datareader"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <config.yaml> [max_records_to_show]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s examples/stream_config.yaml 10\n", os.Args[0])
		os.Exit(1)
	}
	
	configPath := os.Args[1]
	maxToShow := 10
	if len(os.Args) > 2 {
		var err error
		maxToShow, err = parseMaxRecords(os.Args[2])
		if err != nil {
			log.Fatalf("Invalid max_records_to_show: %v", err)
		}
	}
	
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// Create data reader (which will be a stream generator in this case)
	reader, err := datareader.New(cfg.Source)
	if err != nil {
		log.Fatalf("Failed to create data reader: %v", err)
	}
	defer reader.Close()
	
	fmt.Printf("Stream generator started with config: %s\n", configPath)
	fmt.Printf("Showing first %d records...\n\n", maxToShow)
	
	start := time.Now()
	recordCount := 0
	
	// Read and display records
	for recordCount < maxToShow {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				fmt.Printf("\nReached end of stream after %d records.\n", recordCount)
				break
			}
			log.Fatalf("Error reading record: %v", err)
		}
		
		recordCount++
		
		// Pretty print the record as JSON
		jsonData, err := json.MarshalIndent(record, "", "  ")
		if err != nil {
			log.Printf("Error marshaling record %d: %v", recordCount, err)
			continue
		}
		
		fmt.Printf("Record %d:\n%s\n\n", recordCount, string(jsonData))
	}
	
	elapsed := time.Since(start)
	fmt.Printf("Generated and read %d records in %v\n", recordCount, elapsed)
	
	if recordCount > 0 {
		rate := float64(recordCount) / elapsed.Seconds()
		fmt.Printf("Average rate: %.2f records/second\n", rate)
	}
}

func parseMaxRecords(s string) (int, error) {
	var max int
	_, err := fmt.Sscanf(s, "%d", &max)
	if err != nil {
		return 0, err
	}
	if max < 1 {
		return 0, fmt.Errorf("max_records_to_show must be positive")
	}
	return max, nil
}