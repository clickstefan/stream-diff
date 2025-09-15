package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config defines the structure of the user-provided YAML configuration file.
type Config struct {
	Source          Source           `yaml:"source"`
	PatternDetection *PatternDetection `yaml:"pattern_detection,omitempty"`
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

// PatternDetection holds configuration for AI-powered pattern detection.
type PatternDetection struct {
	Enabled bool   `yaml:"enabled"`
	Mode    string `yaml:"mode"` // "offline" or "online"
	
	// Offline mode configuration
	OfflineModel *OfflineModelConfig `yaml:"offline_model,omitempty"`
	
	// Online mode configuration (Claude/Anthropic)
	OnlineAPI *OnlineAPIConfig `yaml:"online_api,omitempty"`
}

// OfflineModelConfig holds configuration for embedded AI model.
type OfflineModelConfig struct {
	ModelPath string `yaml:"model_path,omitempty"` // Path to local model file
	ModelType string `yaml:"model_type,omitempty"` // Type of model (e.g., "onnx", "tflite")
}

// OnlineAPIConfig holds configuration for online AI services.
type OnlineAPIConfig struct {
	Provider string `yaml:"provider"` // "claude" or "anthropic"
	APIKey   string `yaml:"api_key"`
	Model    string `yaml:"model,omitempty"` // Model version to use
	Endpoint string `yaml:"endpoint,omitempty"` // Custom endpoint if needed
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
