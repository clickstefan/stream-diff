package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config defines the structure of the user-provided YAML configuration file.
type Config struct {
	Source Source `yaml:"source"`
}

// RunConfig defines the configuration for running stream comparisons with periodic reporting.
type RunConfig struct {
	Source1 Source        `yaml:"source1"`
	Source2 Source        `yaml:"source2"`
	Output  OutputConfig  `yaml:"output,omitempty"`
	Periodic PeriodicConfig `yaml:"periodic,omitempty"`
}

// OutputConfig defines output settings for comparison results.
type OutputConfig struct {
	FinalReport   string `yaml:"final_report,omitempty"`
	PeriodicReports string `yaml:"periodic_reports,omitempty"` // Directory for periodic reports
}

// PeriodicConfig defines settings for periodic diff reporting.
type PeriodicConfig struct {
	Enabled       bool `yaml:"enabled"`
	TimeInterval  int  `yaml:"time_interval_seconds,omitempty"`  // Time interval in seconds
	RecordInterval int  `yaml:"record_interval,omitempty"`       // Record count interval
}

// Source defines the data source configuration.
type Source struct {
	Type         string        `yaml:"type"`
	Path         string        `yaml:"path"`
	ParserConfig *ParserConfig `yaml:"parser_config,omitempty"`
	Sampler      *Sampler      `yaml:"sampler,omitempty"`
}

// ParserConfig holds optional configuration for the data parser.
type ParserConfig struct {
	JSONInString bool `yaml:"json_in_string"`
}

// Sampler holds optional configuration for the schema generation sampler.
type Sampler struct {
	SampleSize int `yaml:"sample_size"`
}

// Load reads a YAML configuration file from the given path and returns a Config struct.
func Load(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml from %s: %w", filePath, err)
	}

	return &cfg, nil
}

// LoadRunConfig reads a YAML run configuration file from the given path and returns a RunConfig struct.
func LoadRunConfig(filePath string) (*RunConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read run config file %s: %w", filePath, err)
	}

	var cfg RunConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml from %s: %w", filePath, err)
	}

	// Set defaults for periodic configuration
	if cfg.Periodic.TimeInterval == 0 && cfg.Periodic.RecordInterval == 0 {
		cfg.Periodic.TimeInterval = 30 // Default to 30 seconds
		cfg.Periodic.RecordInterval = 1000 // Default to 1000 records
	}

	return &cfg, nil
}
