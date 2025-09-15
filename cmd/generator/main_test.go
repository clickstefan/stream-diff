package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCLIGenerator(t *testing.T) {
	// Build the CLI tool first
	buildCmd := exec.Command("make", "build")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI tool: %v", err)
	}
	
	tests := []struct {
		name     string
		args     []string
		wantRows int
		format   string
	}{
		{
			name:     "CSV with header",
			args:     []string{"-format", "csv", "-count", "5", "-header"},
			wantRows: 6, // 5 data rows + 1 header
			format:   "csv",
		},
		{
			name:     "JSONL format",
			args:     []string{"-format", "jsonl", "-count", "3"},
			wantRows: 3,
			format:   "jsonl",
		},
		{
			name:     "Proto format",
			args:     []string{"-format", "proto", "-count", "2"},
			wantRows: 2,
			format:   "proto",
		},
		{
			name:     "With schema file",
			args:     []string{"-schema", "examples/user_schema.yaml", "-count", "4", "-format", "jsonl"},
			wantRows: 4,
			format:   "jsonl",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./bin/stream-generator", tt.args...)
			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			
			err := cmd.Run()
			if err != nil {
				t.Fatalf("Command failed: %v", err)
			}
			
			output := stdout.String()
			lines := strings.Split(strings.TrimSpace(output), "\n")
			
			if len(lines) != tt.wantRows {
				t.Errorf("Expected %d rows, got %d", tt.wantRows, len(lines))
			}
			
			// Verify format-specific content
			switch tt.format {
			case "csv":
				// First line should be header if requested
				if strings.Contains(strings.Join(tt.args, " "), "-header") {
					if !strings.Contains(lines[0], ",") {
						t.Errorf("CSV header row should contain commas")
					}
				}
			case "jsonl", "proto":
				// Each line should be valid JSON
				for i, line := range lines {
					if line == "" {
						continue
					}
					if !strings.HasPrefix(line, "{") || !strings.HasSuffix(line, "}") {
						t.Errorf("Line %d should be JSON object: %s", i, line)
					}
				}
			}
		})
	}
}

func TestRealWorldSchemas(t *testing.T) {
	schemas := []string{
		"examples/schemas/ecommerce_orders.yaml",
		"examples/schemas/kafka_events.yaml",
		"examples/schemas/app_logs.yaml",
		"examples/schemas/iot_sensors.yaml",
		"examples/schemas/financial_transactions.yaml",
	}
	
	for _, schema := range schemas {
		t.Run(schema, func(t *testing.T) {
			// Check if schema file exists
			if _, err := os.Stat(schema); os.IsNotExist(err) {
				t.Skipf("Schema file %s does not exist", schema)
			}
			
			// Test generation with the schema
			cmd := exec.Command("./bin/stream-generator", "-schema", schema, "-count", "2", "-format", "jsonl")
			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			
			err := cmd.Run()
			if err != nil {
				t.Fatalf("Failed to generate data with schema %s: %v", schema, err)
			}
			
			output := stdout.String()
			lines := strings.Split(strings.TrimSpace(output), "\n")
			
			if len(lines) != 2 {
				t.Errorf("Expected 2 lines, got %d for schema %s", len(lines), schema)
			}
			
			// Verify each line is valid JSON
			for i, line := range lines {
				if !strings.HasPrefix(line, "{") || !strings.HasSuffix(line, "}") {
					t.Errorf("Line %d is not valid JSON for schema %s: %s", i, schema, line)
				}
			}
		})
	}
}

func TestPerformance(t *testing.T) {
	// Test that we can generate a reasonable number of records quickly
	cmd := exec.Command("./bin/stream-generator", "-count", "1000", "-format", "jsonl")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Performance test failed: %v", err)
	}
	
	output := stdout.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	if len(lines) != 1000 {
		t.Errorf("Expected 1000 lines, got %d", len(lines))
	}
}