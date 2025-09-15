package main

import (
	"data-comparator/internal/pkg/config"
	"data-comparator/internal/pkg/datareader"
	"data-comparator/internal/pkg/schema"
	"fmt"
	"log"

	"gopkg.in/yaml.v3"
)

func main() {
	// Demo 1: Basic schema generation (without pattern detection)
	fmt.Println("=== Demo 1: Basic Schema Generation ===")
	demoBasicSchema()

	fmt.Println("\n=== Demo 2: Schema Generation with AI Pattern Detection (Offline Mode) ===")
	demoPatternDetection()
}

func demoBasicSchema() {
	cfg := &config.Config{
		Source: config.Source{
			Type: "csv",
			Path: "testdata/testcase1_simple_csv/source1.csv",
		},
	}

	reader, err := datareader.New(cfg.Source)
	if err != nil {
		log.Fatalf("Failed to create data reader: %v", err)
	}
	defer reader.Close()

	schema, err := schema.Generate(reader, nil)
	if err != nil {
		log.Fatalf("Failed to generate schema: %v", err)
	}

	output, _ := yaml.Marshal(schema)
	fmt.Printf("Basic Schema:\n%s\n", output)
}

func demoPatternDetection() {
	cfg := &config.Config{
		Source: config.Source{
			Type: "csv",
			Path: "testdata/testcase1_simple_csv/source1.csv",
		},
		PatternDetection: &config.PatternDetection{
			Enabled: true,
			Mode:    "offline",
		},
	}

	reader, err := datareader.New(cfg.Source)
	if err != nil {
		log.Fatalf("Failed to create data reader: %v", err)
	}
	defer reader.Close()

	schema, err := schema.GenerateWithPatternDetection(reader, nil, cfg.PatternDetection)
	if err != nil {
		log.Fatalf("Failed to generate schema with pattern detection: %v", err)
	}

	output, _ := yaml.Marshal(schema)
	fmt.Printf("Schema with AI Pattern Detection:\n%s\n", output)
}