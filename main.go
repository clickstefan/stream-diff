package main

import (
	"data-comparator/internal/pkg/config"
	"data-comparator/internal/pkg/datareader"
	"data-comparator/internal/pkg/schema"
	"flag"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

func main() {
	var (
		configPath1 = flag.String("config1", "", "Path to first configuration file")
		configPath2 = flag.String("config2", "", "Path to second configuration file")
		outputPath  = flag.String("output", "", "Path to output file (optional, prints to stdout if not provided)")
		help        = flag.Bool("help", false, "Show help")
		version     = flag.Bool("version", false, "Show version")
	)
	flag.Parse()

	if *help {
		fmt.Println("Data Stream Comparator")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  data-comparator -config1 <path> -config2 <path> [-output <path>]")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		return
	}

	if *version {
		fmt.Println("data-comparator version: dev")
		return
	}

	if *configPath1 == "" || *configPath2 == "" {
		fmt.Fprintf(os.Stderr, "Error: Both -config1 and -config2 are required\n")
		fmt.Fprintf(os.Stderr, "Use -help for usage information\n")
		os.Exit(1)
	}

	// Load configurations
	config1, err := config.Load(*configPath1)
	if err != nil {
		log.Fatalf("Failed to load config1: %v", err)
	}

	config2, err := config.Load(*configPath2)
	if err != nil {
		log.Fatalf("Failed to load config2: %v", err)
	}

	// Create data readers
	reader1, err := datareader.New(config1.Source)
	if err != nil {
		log.Fatalf("Failed to create reader for config1: %v", err)
	}

	reader2, err := datareader.New(config2.Source)
	if err != nil {
		log.Fatalf("Failed to create reader for config2: %v", err)
	}

	// Generate schemas
	schema1, err := schema.Generate(reader1, config1.Source.Sampler)
	if err != nil {
		log.Fatalf("Failed to generate schema for config1: %v", err)
	}

	schema2, err := schema.Generate(reader2, config2.Source.Sampler)
	if err != nil {
		log.Fatalf("Failed to generate schema for config2: %v", err)
	}

	// Create comparison result
	result := map[string]interface{}{
		"source1_schema": schema1,
		"source2_schema": schema2,
	}

	// Output result
	yamlData, err := yaml.Marshal(result)
	if err != nil {
		log.Fatalf("Failed to marshal result to YAML: %v", err)
	}

	if *outputPath != "" {
		err = os.WriteFile(*outputPath, yamlData, 0644)
		if err != nil {
			log.Fatalf("Failed to write to file %s: %v", *outputPath, err)
		}
		fmt.Printf("Comparison result written to %s\n", *outputPath)
	} else {
		fmt.Print(string(yamlData))
	}
}