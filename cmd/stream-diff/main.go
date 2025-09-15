package main

import (
	"data-comparator/internal/pkg/comparator"
	"data-comparator/internal/pkg/config"
	"data-comparator/internal/pkg/datareader"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

func main() {
	var (
		runConfigPath = flag.String("config", "runConfig.yaml", "Path to run configuration file")
		source1Path   = flag.String("source1", "", "Path to first source configuration file")
		source2Path   = flag.String("source2", "", "Path to second source configuration file")
		keyField      = flag.String("key", "", "Key field for record comparison")
		timeInterval  = flag.Int("time-interval", 0, "Time interval for periodic reports (seconds)")
		recordInterval = flag.Int("record-interval", 0, "Record interval for periodic reports")
		enablePeriodic = flag.Bool("enable-periodic", false, "Enable periodic reporting")
		outputDir     = flag.String("output-dir", ".", "Directory for output files")
	)
	flag.Parse()

	// Load configuration
	var runConfig *config.RunConfig
	var err error

	if *source1Path != "" && *source2Path != "" {
		// Create run config from command line arguments
		runConfig = &config.RunConfig{
			Source1: config.Source{Type: "csv", Path: *source1Path}, // Default to CSV
			Source2: config.Source{Type: "csv", Path: *source2Path}, // Default to CSV
			Output: config.OutputConfig{
				FinalReport:     filepath.Join(*outputDir, "final_report.yaml"),
				PeriodicReports: filepath.Join(*outputDir, "periodic_reports"),
			},
			Periodic: config.PeriodicConfig{
				Enabled:        *enablePeriodic,
				TimeInterval:   *timeInterval,
				RecordInterval: *recordInterval,
			},
		}
	} else {
		// Load from config file
		runConfig, err = config.LoadRunConfig(*runConfigPath)
		if err != nil {
			log.Fatalf("Failed to load run config: %v", err)
		}
	}

	// Override key field if provided
	if *keyField != "" {
		// Key field will be passed to comparator
	} else {
		// Try to detect key field or use a default
		*keyField = "id" // Default key field
	}

	fmt.Printf("Starting stream comparison...\n")
	fmt.Printf("Source 1: %s\n", runConfig.Source1.Path)
	fmt.Printf("Source 2: %s\n", runConfig.Source2.Path)
	fmt.Printf("Key field: %s\n", *keyField)
	if runConfig.Periodic.Enabled {
		fmt.Printf("Periodic reporting enabled - Time: %ds, Records: %d\n", 
			runConfig.Periodic.TimeInterval, runConfig.Periodic.RecordInterval)
	}

	// Create data readers
	reader1, err := datareader.New(runConfig.Source1)
	if err != nil {
		log.Fatalf("Failed to create reader for source1: %v", err)
	}
	defer reader1.Close()

	reader2, err := datareader.New(runConfig.Source2)
	if err != nil {
		log.Fatalf("Failed to create reader for source2: %v", err)
	}
	defer reader2.Close()

	// Create periodic reports directory if needed
	if runConfig.Periodic.Enabled && runConfig.Output.PeriodicReports != "" {
		if err := os.MkdirAll(runConfig.Output.PeriodicReports, 0755); err != nil {
			log.Fatalf("Failed to create periodic reports directory: %v", err)
		}
	}

	// Create periodic diff callback
	periodicDiffCallback := func(result comparator.ComparisonResult) error {
		fmt.Printf("[PERIODIC] %s - Records: %d, Matching: %d, Identical: %d, Diffs: %d\n",
			result.Timestamp.Format("15:04:05"),
			result.RecordsProcessed,
			result.MatchingKeys,
			result.IdenticalRows,
			len(result.ValueDiffs))

		// Save periodic report if configured
		if runConfig.Output.PeriodicReports != "" {
			filename := fmt.Sprintf("periodic_report_%s.yaml", 
				result.Timestamp.Format("20060102_150405"))
			filePath := filepath.Join(runConfig.Output.PeriodicReports, filename)
			
			if err := saveReportToFile(result, filePath); err != nil {
				return fmt.Errorf("failed to save periodic report: %w", err)
			}
			fmt.Printf("  Saved periodic report: %s\n", filePath)
		}

		return nil
	}

	// Create stream comparator
	sc := comparator.NewStreamComparator(
		reader1,
		reader2,
		runConfig.Periodic,
		*keyField,
		periodicDiffCallback,
	)

	// Perform comparison
	fmt.Printf("Starting comparison...\n")
	startTime := time.Now()
	
	finalResult, err := sc.Compare()
	if err != nil {
		log.Fatalf("Comparison failed: %v", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("\nComparison completed in %v\n", duration)

	// Print final summary
	fmt.Printf("\nFinal Results:\n")
	fmt.Printf("  Records processed: %d\n", finalResult.RecordsProcessed)
	fmt.Printf("  Source1 records: %d\n", finalResult.Source1Records)
	fmt.Printf("  Source2 records: %d\n", finalResult.Source2Records)
	fmt.Printf("  Matching keys: %d\n", finalResult.MatchingKeys)
	fmt.Printf("  Identical rows: %d\n", finalResult.IdenticalRows)
	fmt.Printf("  Value differences: %d records\n", len(finalResult.ValueDiffs))
	fmt.Printf("  Keys only in source1: %d\n", len(finalResult.KeysOnlyInSource1))
	fmt.Printf("  Keys only in source2: %d\n", len(finalResult.KeysOnlyInSource2))

	// Save final report
	finalReportPath := runConfig.Output.FinalReport
	if finalReportPath == "" {
		finalReportPath = "final_report.yaml"
	}
	
	if err := saveReportToFile(*finalResult, finalReportPath); err != nil {
		log.Fatalf("Failed to save final report: %v", err)
	}
	fmt.Printf("\nFinal report saved to: %s\n", finalReportPath)
}

// saveReportToFile saves a comparison result to a YAML file.
func saveReportToFile(result interface{}, filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Marshal to YAML
	data, err := yaml.Marshal(result)
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(filePath, data, 0644)
}