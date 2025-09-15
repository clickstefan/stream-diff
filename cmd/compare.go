package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"data-comparator/internal/pkg/config"
	"data-comparator/internal/pkg/datareader"
	"data-comparator/internal/pkg/schema"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	outputFile   string
	sampleSize   int
	schemaOnly   bool
	outputFormat string
)

// compareCmd represents the compare command
var compareCmd = &cobra.Command{
	Use:   "compare <config1.yaml> <config2.yaml>",
	Short: "Compare two data streams and generate a detailed report",
	Long: `Compare two data streams from various sources and generate a comprehensive
analysis report including schema differences, value mismatches, and statistics.

The comparison process includes:
  • Schema detection and analysis for both sources
  • Record-by-record comparison using configurable key fields
  • Statistical analysis of field distributions and patterns  
  • Detection of data quality issues and anomalies
  • AI-powered insights and recommendations for data discrepancies

Examples:
  # Basic comparison
  stream-diff compare source1.yaml source2.yaml

  # Generate schema only (faster for large datasets)
  stream-diff compare --schema-only source1.yaml source2.yaml

  # Limit sample size for quick analysis
  stream-diff compare --sample-size 1000 source1.yaml source2.yaml

  # Save results to specific file
  stream-diff compare --output report.yaml source1.yaml source2.yaml`,
	Args: cobra.ExactArgs(2),
	RunE: runCompare,
}

func init() {
	rootCmd.AddCommand(compareCmd)

	// Comparison-specific flags
	compareCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file for comparison report (default: stdout)")
	compareCmd.Flags().IntVar(&sampleSize, "sample-size", 0, "limit sample size for schema generation (0 = no limit)")
	compareCmd.Flags().BoolVar(&schemaOnly, "schema-only", false, "generate schemas only, skip full comparison")
	compareCmd.Flags().StringVar(&outputFormat, "format", "yaml", "output format: yaml, json")

	// Add AI-enhanced help
	compareCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Println(aiEnhancedHelp(cmd))
		cmd.Parent().HelpFunc()(cmd, args)
	})
}

func runCompare(cmd *cobra.Command, args []string) error {
	config1Path := args[0]
	config2Path := args[1]

	log.Info().
		Str("config1", config1Path).
		Str("config2", config2Path).
		Bool("schema_only", schemaOnly).
		Int("sample_size", sampleSize).
		Msg("Starting comparison")

	// Load configurations
	cfg1, err := config.Load(config1Path)
	if err != nil {
		return fmt.Errorf("failed to load config1 %s: %w", config1Path, err)
	}

	cfg2, err := config.Load(config2Path)
	if err != nil {
		return fmt.Errorf("failed to load config2 %s: %w", config2Path, err)
	}

	// Validate configurations
	if err := validateConfig(cfg1, config1Path); err != nil {
		return fmt.Errorf("config1 validation failed: %w", err)
	}

	if err := validateConfig(cfg2, config2Path); err != nil {
		return fmt.Errorf("config2 validation failed: %w", err)
	}

	// Apply sample size override if specified
	if sampleSize > 0 {
		if cfg1.Source.Sampler == nil {
			cfg1.Source.Sampler = &config.Sampler{}
		}
		if cfg2.Source.Sampler == nil {
			cfg2.Source.Sampler = &config.Sampler{}
		}
		cfg1.Source.Sampler.SampleSize = sampleSize
		cfg2.Source.Sampler.SampleSize = sampleSize
		log.Info().Int("sample_size", sampleSize).Msg("Applied sample size override")
	}

	// Generate schemas
	log.Info().Msg("Generating schema for source 1")
	schema1, err := generateSchema(cfg1, "source1")
	if err != nil {
		return fmt.Errorf("failed to generate schema for source1: %w", err)
	}

	log.Info().Msg("Generating schema for source 2")
	schema2, err := generateSchema(cfg2, "source2")
	if err != nil {
		return fmt.Errorf("failed to generate schema for source2: %w", err)
	}

	// Create comparison result
	result := ComparisonResult{
		Metadata: ComparisonMetadata{
			Timestamp:    time.Now(),
			Source1Path:  config1Path,
			Source2Path:  config2Path,
			SchemaOnly:   schemaOnly,
			SampleSize:   sampleSize,
			ToolVersion:  "1.0.0", // TODO: Get from build info
		},
		Schema1: schema1,
		Schema2: schema2,
	}

	if !schemaOnly {
		log.Info().Msg("Performing full comparison (not implemented yet)")
		// TODO: Implement full comparison logic
		result.Summary = &ComparisonSummary{
			Status: "schema_generated",
			Notes:  "Full comparison not yet implemented - schemas generated successfully",
		}
	}

	// Add AI-powered insights
	result.AIInsights = generateAIInsights(schema1, schema2)

	// Output results
	return outputResult(result)
}

// validateConfig validates a configuration for common issues
func validateConfig(cfg *config.Config, configPath string) error {
	// Check if source file exists - try both relative to config and current dir
	sourcePath := cfg.Source.Path
	var resolvedPath string
	
	if filepath.IsAbs(sourcePath) {
		resolvedPath = sourcePath
	} else {
		// First try relative to config file directory
		configDir := filepath.Dir(configPath)
		configRelativePath := filepath.Join(configDir, sourcePath)
		
		if _, err := os.Stat(configRelativePath); err == nil {
			resolvedPath = configRelativePath
		} else if _, err := os.Stat(sourcePath); err == nil {
			// Try relative to current working directory
			resolvedPath = sourcePath
		} else {
			return fmt.Errorf("source file does not exist: %s (tried %s and %s)", sourcePath, configRelativePath, sourcePath)
		}
	}

	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", resolvedPath)
	}

	// Validate source type
	supportedTypes := map[string]bool{
		"csv":  true,
		"json": true,
	}

	if !supportedTypes[cfg.Source.Type] {
		return fmt.Errorf("unsupported source type: %s (supported: csv, json)", cfg.Source.Type)
	}

	return nil
}

// generateSchema generates a schema for the given configuration
func generateSchema(cfg *config.Config, sourceName string) (*schema.Schema, error) {
	// No need to adjust path here - datareader should handle relative paths correctly
	reader, err := datareader.New(cfg.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to create data reader for %s: %w", sourceName, err)
	}
	defer reader.Close()

	schema, err := schema.Generate(reader, cfg.Source.Sampler)
	if err != nil {
		return nil, fmt.Errorf("failed to generate schema for %s: %w", sourceName, err)
	}

	log.Info().
		Str("source", sourceName).
		Int("fields", len(schema.Fields)).
		Str("key", schema.Key).
		Msg("Schema generated")

	return schema, nil
}

// generateAIInsights creates AI-powered insights from schema comparison
func generateAIInsights(schema1, schema2 *schema.Schema) []AIInsight {
	var insights []AIInsight

	// Schema structure comparison
	if len(schema1.Fields) != len(schema2.Fields) {
		insights = append(insights, AIInsight{
			Type:       "schema_structure",
			Severity:   "medium",
			Message:    fmt.Sprintf("Field count differs: source1 has %d fields, source2 has %d fields", len(schema1.Fields), len(schema2.Fields)),
			Suggestion: "Review field mappings and consider if this difference is expected. Missing fields may indicate data quality issues.",
		})
	}

	// Key field comparison
	if schema1.Key != schema2.Key {
		insights = append(insights, AIInsight{
			Type:       "key_mismatch",
			Severity:   "high",
			Message:    fmt.Sprintf("Different key fields: source1 uses '%s', source2 uses '%s'", schema1.Key, schema2.Key),
			Suggestion: "Key field mismatch will prevent proper record comparison. Ensure both sources use the same identifier field.",
		})
	}

	// Field type mismatches
	for fieldName, field1 := range schema1.Fields {
		if field2, exists := schema2.Fields[fieldName]; exists {
			if field1.Type != field2.Type {
				insights = append(insights, AIInsight{
					Type:       "type_mismatch",
					Severity:   "medium",
					Message:    fmt.Sprintf("Field '%s' type mismatch: source1=%s, source2=%s", fieldName, field1.Type, field2.Type),
					Suggestion: "Type mismatches may indicate data format differences or quality issues. Consider data transformation or validation rules.",
				})
			}
		}
	}

	// Missing fields
	for fieldName := range schema1.Fields {
		if _, exists := schema2.Fields[fieldName]; !exists {
			insights = append(insights, AIInsight{
				Type:       "missing_field",
				Severity:   "medium",
				Message:    fmt.Sprintf("Field '%s' exists in source1 but missing in source2", fieldName),
				Suggestion: "Missing fields may indicate incomplete data migration or different data collection processes.",
			})
		}
	}

	for fieldName := range schema2.Fields {
		if _, exists := schema1.Fields[fieldName]; !exists {
			insights = append(insights, AIInsight{
				Type:       "extra_field",
				Severity:   "low",
				Message:    fmt.Sprintf("Field '%s' exists in source2 but missing in source1", fieldName),
				Suggestion: "Additional fields might represent new data features or different schema versions.",
			})
		}
	}

	if len(insights) == 0 {
		insights = append(insights, AIInsight{
			Type:       "schema_compatible",
			Severity:   "info",
			Message:    "Schemas appear compatible with same fields and types",
			Suggestion: "Good data consistency detected. Proceed with full comparison for detailed value analysis.",
		})
	}

	return insights
}

// outputResult writes the comparison result to the specified output
func outputResult(result ComparisonResult) error {
	var output []byte
	var err error

	switch outputFormat {
	case "yaml":
		output, err = yaml.Marshal(result)
	case "json":
		// TODO: Add JSON marshaling when needed
		return fmt.Errorf("json output format not yet implemented")
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	// Write to file or stdout
	if outputFile != "" {
		log.Info().Str("file", outputFile).Msg("Writing results to file")
		return os.WriteFile(outputFile, output, 0644)
	}

	fmt.Print(string(output))
	return nil
}

// Data structures for comparison results

type ComparisonResult struct {
	Metadata   ComparisonMetadata `yaml:"metadata"`
	Schema1    *schema.Schema     `yaml:"schema1"`
	Schema2    *schema.Schema     `yaml:"schema2"`
	Summary    *ComparisonSummary `yaml:"summary,omitempty"`
	AIInsights []AIInsight        `yaml:"ai_insights,omitempty"`
}

type ComparisonMetadata struct {
	Timestamp    time.Time `yaml:"timestamp"`
	Source1Path  string    `yaml:"source1_path"`
	Source2Path  string    `yaml:"source2_path"`
	SchemaOnly   bool      `yaml:"schema_only"`
	SampleSize   int       `yaml:"sample_size"`
	ToolVersion  string    `yaml:"tool_version"`
}

type ComparisonSummary struct {
	Status string `yaml:"status"`
	Notes  string `yaml:"notes"`
}

type AIInsight struct {
	Type       string `yaml:"type"`
	Severity   string `yaml:"severity"`
	Message    string `yaml:"message"`
	Suggestion string `yaml:"suggestion"`
}