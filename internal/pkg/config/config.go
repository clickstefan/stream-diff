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

// Source defines the data source configuration.
type Source struct {
	Type         string        `yaml:"type"`
	Path         string        `yaml:"path,omitempty"`
	ParserConfig *ParserConfig `yaml:"parser_config,omitempty"`
	Sampler      *Sampler      `yaml:"sampler,omitempty"`
	StreamGenerator *StreamGeneratorConfig `yaml:"stream_generator,omitempty"`
}

// ParserConfig holds optional configuration for the data parser.
type ParserConfig struct {
	JSONInString bool `yaml:"json_in_string"`
}

// Sampler holds optional configuration for the schema generation sampler.
type Sampler struct {
	SampleSize int `yaml:"sample_size"`
}

// StreamGeneratorConfig holds configuration for the stream generator.
type StreamGeneratorConfig struct {
	// SchemaPath points to a schema file to use for generation
	SchemaPath string `yaml:"schema_path,omitempty"`
	
	// MaxRecords limits the total number of records generated (0 = unlimited)
	MaxRecords int64 `yaml:"max_records,omitempty"`
	
	// RecordsPerSecond controls the generation rate (0 = no rate limiting)
	RecordsPerSecond float64 `yaml:"records_per_second,omitempty"`
	
	// BufferSize controls the internal channel buffer size for backpressure
	BufferSize int `yaml:"buffer_size,omitempty"`
	
	// Seed for random number generator (0 = use current time)
	Seed int64 `yaml:"seed,omitempty"`
	
	// Patterns for generating realistic data
	DataPatterns map[string]DataPattern `yaml:"data_patterns,omitempty"`
}

// DataPattern defines how to generate realistic data for a specific field pattern.
type DataPattern struct {
	// Type can be "list", "range", "format", "expression"
	Type string `yaml:"type"`
	
	// Values for list-type patterns
	Values []interface{} `yaml:"values,omitempty"`
	
	// Min/Max for range-type patterns
	Min interface{} `yaml:"min,omitempty"`
	Max interface{} `yaml:"max,omitempty"`
	
	// Format string for format-type patterns (e.g., email, phone)
	Format string `yaml:"format,omitempty"`
	
	// Expression for expression-type patterns
	Expression string `yaml:"expression,omitempty"`
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
