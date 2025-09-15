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
