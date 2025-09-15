package config

import (
	"testing"
)

func TestLoad(t *testing.T) {
	// Use one of the test configs we created.
	// The path needs to be relative to the package directory where `go test` is run.
	filePath := "../../../testdata/testcase3_csv_with_json/config1.yaml"

	cfg, err := Load(filePath)
	if err != nil {
		t.Fatalf("Load() error = %v, wantErr %v", err, false)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	// Check some values to make sure parsing worked.
	if cfg.Source.Type != "csv" {
		t.Errorf("Source.Type got = %v, want %v", cfg.Source.Type, "csv")
	}
	if cfg.Source.Path != "testdata/testcase3_csv_with_json/source1.csv" {
		t.Errorf("Source.Path got = %v, want %v", cfg.Source.Path, "testdata/testcase3_csv_with_json/source1.csv")
	}
	if cfg.Source.ParserConfig == nil {
		t.Fatal("Source.ParserConfig is nil")
	}
	if !cfg.Source.ParserConfig.JSONInString {
		t.Errorf("Source.ParserConfig.JSONInString got = %v, want %v", cfg.Source.ParserConfig.JSONInString, true)
	}
}
