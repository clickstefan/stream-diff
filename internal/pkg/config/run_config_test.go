package config

import (
	"os"
	"testing"
)

func TestLoadRunConfig(t *testing.T) {
	// Create a test config file content
	yamlContent := `
source1:
  type: csv
  path: testdata/source1.csv
source2:
  type: csv
  path: testdata/source2.csv
output:
  final_report: final_report.yaml
  periodic_reports: periodic_reports
periodic:
  enabled: true
  time_interval_seconds: 30
  record_interval: 1000
`
	// Create temporary file
	tmpFile := "/tmp/test_run_config.yaml"
	err := writeStringToFile(tmpFile, yamlContent)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test loading
	runConfig, err := LoadRunConfig(tmpFile)
	if err != nil {
		t.Fatalf("LoadRunConfig() error = %v", err)
	}

	if runConfig == nil {
		t.Fatal("LoadRunConfig() returned nil config")
	}

	// Verify source1
	if runConfig.Source1.Type != "csv" {
		t.Errorf("Source1.Type got = %v, want %v", runConfig.Source1.Type, "csv")
	}
	if runConfig.Source1.Path != "testdata/source1.csv" {
		t.Errorf("Source1.Path got = %v, want %v", runConfig.Source1.Path, "testdata/source1.csv")
	}

	// Verify source2
	if runConfig.Source2.Type != "csv" {
		t.Errorf("Source2.Type got = %v, want %v", runConfig.Source2.Type, "csv")
	}
	if runConfig.Source2.Path != "testdata/source2.csv" {
		t.Errorf("Source2.Path got = %v, want %v", runConfig.Source2.Path, "testdata/source2.csv")
	}

	// Verify output config
	if runConfig.Output.FinalReport != "final_report.yaml" {
		t.Errorf("Output.FinalReport got = %v, want %v", runConfig.Output.FinalReport, "final_report.yaml")
	}
	if runConfig.Output.PeriodicReports != "periodic_reports" {
		t.Errorf("Output.PeriodicReports got = %v, want %v", runConfig.Output.PeriodicReports, "periodic_reports")
	}

	// Verify periodic config
	if !runConfig.Periodic.Enabled {
		t.Errorf("Periodic.Enabled got = %v, want %v", runConfig.Periodic.Enabled, true)
	}
	if runConfig.Periodic.TimeInterval != 30 {
		t.Errorf("Periodic.TimeInterval got = %v, want %v", runConfig.Periodic.TimeInterval, 30)
	}
	if runConfig.Periodic.RecordInterval != 1000 {
		t.Errorf("Periodic.RecordInterval got = %v, want %v", runConfig.Periodic.RecordInterval, 1000)
	}
}

func TestLoadRunConfigWithDefaults(t *testing.T) {
	// Create a minimal test config file content (no periodic config)
	yamlContent := `
source1:
  type: csv
  path: testdata/source1.csv
source2:
  type: json
  path: testdata/source2.json
`
	// Create temporary file
	tmpFile := "/tmp/test_run_config_defaults.yaml"
	err := writeStringToFile(tmpFile, yamlContent)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test loading
	runConfig, err := LoadRunConfig(tmpFile)
	if err != nil {
		t.Fatalf("LoadRunConfig() error = %v", err)
	}

	// Verify defaults are applied
	if runConfig.Periodic.TimeInterval != 30 {
		t.Errorf("Default Periodic.TimeInterval got = %v, want %v", runConfig.Periodic.TimeInterval, 30)
	}
	if runConfig.Periodic.RecordInterval != 1000 {
		t.Errorf("Default Periodic.RecordInterval got = %v, want %v", runConfig.Periodic.RecordInterval, 1000)
	}
}

// Helper function to write string to file
func writeStringToFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}