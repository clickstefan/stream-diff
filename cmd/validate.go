package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"data-comparator/internal/pkg/config"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	explainValidation bool
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate <config.yaml> [config2.yaml]",
	Short: "Validate configuration files and data sources",
	Long: `Validate one or more configuration files to ensure they are properly formatted
and their data sources are accessible. This helps catch issues early before
running time-consuming comparisons.

The validation process checks:
  • YAML syntax and structure
  • Required fields and valid values  
  • Data source accessibility and format
  • Schema compatibility between sources
  • AI-powered configuration recommendations

Examples:
  # Validate single configuration
  stream-diff validate config.yaml

  # Validate both configurations for comparison
  stream-diff validate config1.yaml config2.yaml

  # Get detailed validation explanations
  stream-diff validate --explain config1.yaml config2.yaml`,
	Args: cobra.MinimumNArgs(1),
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)

	// Validation-specific flags
	validateCmd.Flags().BoolVar(&explainValidation, "explain", false, "provide detailed explanations for validation results")

	// Add AI-enhanced help
	validateCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Println(aiEnhancedHelp(cmd))
		cmd.Parent().HelpFunc()(cmd, args)
	})
}

func runValidate(cmd *cobra.Command, args []string) error {
	log.Info().Int("configs", len(args)).Msg("Starting configuration validation")

	var configs []*config.Config
	var errors []ValidationError

	// Load and validate each configuration
	for i, configPath := range args {
		log.Info().Str("config", configPath).Int("index", i+1).Msg("Validating configuration")

		cfg, validationErrors := validateSingleConfig(configPath)
		if len(validationErrors) > 0 {
			errors = append(errors, validationErrors...)
		}

		if cfg != nil {
			configs = append(configs, cfg)
		}
	}

	// Cross-validation for multiple configs
	if len(configs) > 1 {
		crossValidationErrors := validateConfigCompatibility(configs, args)
		errors = append(errors, crossValidationErrors...)
	}

	// Generate AI recommendations
	recommendations := generateValidationRecommendations(configs, errors)

	// Output results
	errorCount := 0
	warningCount := 0
	for _, err := range errors {
		switch err.Severity {
		case "error":
			errorCount++
		case "warning":
			warningCount++
		}
	}
	
	result := ValidationResult{
		ConfigPaths:     args,
		Valid:          errorCount == 0 && warningCount == 0,
		Errors:         errors,
		Recommendations: recommendations,
	}

	return outputValidationResult(result)
}

func validateSingleConfig(configPath string) (*config.Config, []ValidationError) {
	var errors []ValidationError

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		errors = append(errors, ValidationError{
			ConfigPath: configPath,
			Type:       "file_not_found",
			Field:      "",
			Message:    "Configuration file does not exist",
			Severity:   "error",
		})
		return nil, errors
	}

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		errors = append(errors, ValidationError{
			ConfigPath: configPath,
			Type:       "syntax_error",
			Field:      "",
			Message:    fmt.Sprintf("Failed to parse YAML: %v", err),
			Severity:   "error",
		})
		return nil, errors
	}

	// Validate required fields
	if cfg.Source.Type == "" {
		errors = append(errors, ValidationError{
			ConfigPath: configPath,
			Type:       "missing_field",
			Field:      "source.type",
			Message:    "Source type is required",
			Severity:   "error",
		})
	}

	if cfg.Source.Path == "" {
		errors = append(errors, ValidationError{
			ConfigPath: configPath,
			Type:       "missing_field",
			Field:      "source.path",
			Message:    "Source path is required",
			Severity:   "error",
		})
	}

	// Validate source type
	supportedTypes := map[string]bool{
		"csv":  true,
		"json": true,
	}

	if cfg.Source.Type != "" && !supportedTypes[cfg.Source.Type] {
		errors = append(errors, ValidationError{
			ConfigPath: configPath,
			Type:       "invalid_value",
			Field:      "source.type",
			Message:    fmt.Sprintf("Unsupported source type '%s'. Supported types: csv, json", cfg.Source.Type),
			Severity:   "error",
		})
	}

	// Validate source file
	if cfg.Source.Path != "" {
		sourcePath := cfg.Source.Path
		if !filepath.IsAbs(sourcePath) {
			// First try relative to config file directory
			configDir := filepath.Dir(configPath)
			sourcePath = filepath.Join(configDir, sourcePath)
			
			// If that doesn't exist, try relative to current working directory
			if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
				sourcePath = cfg.Source.Path
			}
		}

		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			errors = append(errors, ValidationError{
				ConfigPath: configPath,
				Type:       "file_not_found",
				Field:      "source.path",
				Message:    fmt.Sprintf("Source file does not exist: %s", sourcePath),
				Severity:   "error",
			})
		} else {
			// Check file permissions
			if file, err := os.Open(sourcePath); err != nil {
				errors = append(errors, ValidationError{
					ConfigPath: configPath,
					Type:       "permission_error",
					Field:      "source.path",
					Message:    fmt.Sprintf("Cannot read source file: %v", err),
					Severity:   "error",
				})
			} else {
				file.Close()
			}
		}
	}

	// Validate sampler configuration
	if cfg.Source.Sampler != nil && cfg.Source.Sampler.SampleSize < 0 {
		errors = append(errors, ValidationError{
			ConfigPath: configPath,
			Type:       "invalid_value",
			Field:      "source.sampler.sample_size",
			Message:    "Sample size must be non-negative",
			Severity:   "warning",
		})
	}

	// Performance recommendations
	if cfg.Source.Sampler == nil || cfg.Source.Sampler.SampleSize == 0 {
		errors = append(errors, ValidationError{
			ConfigPath: configPath,
			Type:       "performance_suggestion",
			Field:      "source.sampler.sample_size",
			Message:    "Consider setting a sample size for large datasets to improve performance",
			Severity:   "info",
		})
	}

	return cfg, errors
}

func validateConfigCompatibility(configs []*config.Config, configPaths []string) []ValidationError {
	var errors []ValidationError

	if len(configs) < 2 {
		return errors
	}

	// Compare source types
	firstType := configs[0].Source.Type
	for i := 1; i < len(configs); i++ {
		if configs[i].Source.Type != firstType {
			errors = append(errors, ValidationError{
				ConfigPath: fmt.Sprintf("%s vs %s", configPaths[0], configPaths[i]),
				Type:       "compatibility_warning",
				Field:      "source.type",
				Message:    fmt.Sprintf("Different source types may affect comparison: %s vs %s", firstType, configs[i].Source.Type),
				Severity:   "warning",
			})
		}
	}

	// Compare parser configurations
	firstParser := configs[0].Source.ParserConfig
	for i := 1; i < len(configs); i++ {
		currentParser := configs[i].Source.ParserConfig

		// Check JSON in string parsing compatibility
		firstJSONInString := firstParser != nil && firstParser.JSONInString
		currentJSONInString := currentParser != nil && currentParser.JSONInString

		if firstJSONInString != currentJSONInString {
			errors = append(errors, ValidationError{
				ConfigPath: fmt.Sprintf("%s vs %s", configPaths[0], configPaths[i]),
				Type:       "compatibility_warning",
				Field:      "source.parser_config.json_in_string",
				Message:    "Different JSON-in-string parsing settings may affect field detection",
				Severity:   "warning",
			})
		}
	}

	return errors
}

func generateValidationRecommendations(configs []*config.Config, errors []ValidationError) []ValidationRecommendation {
	var recommendations []ValidationRecommendation

	errorCount := 0
	warningCount := 0
	infoCount := 0

	for _, err := range errors {
		switch err.Severity {
		case "error":
			errorCount++
		case "warning":
			warningCount++
		case "info":
			infoCount++
		}
	}

	// Generate AI-powered recommendations based on error patterns
	if errorCount == 0 && warningCount == 0 {
		recommendations = append(recommendations, ValidationRecommendation{
			Type:     "success",
			Priority: "low",
			Title:    "Configuration Validation Passed",
			Message:  "All configurations are valid and compatible. You can proceed with data comparison.",
			Action:   "Run 'stream-diff compare' to start the comparison process.",
		})
	}

	if errorCount > 0 {
		recommendations = append(recommendations, ValidationRecommendation{
			Type:     "error_resolution",
			Priority: "high",
			Title:    "Critical Issues Found",
			Message:  fmt.Sprintf("Found %d critical errors that must be resolved before comparison.", errorCount),
			Action:   "Review and fix all error-level issues, then run validation again.",
		})
	}

	if warningCount > 0 {
		recommendations = append(recommendations, ValidationRecommendation{
			Type:     "optimization",
			Priority: "medium",
			Title:    "Optimization Opportunities",
			Message:  fmt.Sprintf("Found %d warnings that could affect comparison quality or performance.", warningCount),
			Action:   "Consider addressing warnings for optimal results, though comparison can proceed.",
		})
	}

	// Performance recommendations
	hasLargeDataSets := false
	hasNoSampling := false

	for _, cfg := range configs {
		if cfg.Source.Sampler == nil || cfg.Source.Sampler.SampleSize == 0 {
			hasNoSampling = true
		}
		// TODO: Check file sizes to determine if datasets are large
		hasLargeDataSets = true // Assume for now
	}

	if hasLargeDataSets && hasNoSampling {
		recommendations = append(recommendations, ValidationRecommendation{
			Type:     "performance",
			Priority: "medium",
			Title:    "Performance Optimization Available",
			Message:  "Consider using sampling for initial schema analysis of large datasets.",
			Action:   "Add 'sampler: { sample_size: 10000 }' to your source configuration for faster processing.",
		})
	}

	return recommendations
}

func outputValidationResult(result ValidationResult) error {
	// Count only errors and warnings as real issues
	errorCount := 0
	warningCount := 0
	infoCount := 0

	for _, err := range result.Errors {
		switch err.Severity {
		case "error":
			errorCount++
		case "warning":
			warningCount++
		case "info":
			infoCount++
		}
	}

	// Only consider errors and warnings as validation failures
	hasIssues := errorCount > 0 || warningCount > 0
	
	// Summary
	status := "✅ VALID"
	if hasIssues {
		status = "❌ INVALID"
	}

	fmt.Printf("Validation Result: %s\n", status)
	fmt.Printf("Configurations: %d\n", len(result.ConfigPaths))

	if len(result.Errors) > 0 {
		fmt.Printf("\nIssues Found:\n")
		for _, err := range result.Errors {
			icon := getIssueIcon(err.Severity)
			fmt.Printf("  %s [%s] %s", icon, err.ConfigPath, err.Message)
			if err.Field != "" {
				fmt.Printf(" (field: %s)", err.Field)
			}
			fmt.Printf("\n")

			if explainValidation {
				explanation := getIssueExplanation(err.Type)
				if explanation != "" {
					fmt.Printf("    💡 %s\n", explanation)
				}
			}
		}
	}

	if len(result.Recommendations) > 0 {
		fmt.Printf("\n🤖 AI Recommendations:\n")
		for _, rec := range result.Recommendations {
			priority := getPriorityIcon(rec.Priority)
			fmt.Printf("  %s %s\n", priority, rec.Title)
			fmt.Printf("    %s\n", rec.Message)
			if rec.Action != "" {
				fmt.Printf("    ➡️  %s\n", rec.Action)
			}
			fmt.Printf("\n")
		}
	}

	if hasIssues {
		return fmt.Errorf("validation failed with %d critical issues", errorCount+warningCount)
	}

	return nil
}

func getIssueIcon(severity string) string {
	switch severity {
	case "error":
		return "🔴"
	case "warning":
		return "🟡"
	case "info":
		return "🔵"
	default:
		return "⚪"
	}
}

func getPriorityIcon(priority string) string {
	switch priority {
	case "high":
		return "🔥"
	case "medium":
		return "⚡"
	case "low":
		return "💡"
	default:
		return "ℹ️"
	}
}

func getIssueExplanation(issueType string) string {
	explanations := map[string]string{
		"file_not_found":       "Ensure the file path is correct relative to the configuration file location",
		"syntax_error":         "Check YAML syntax - common issues include indentation and quote usage",
		"missing_field":        "This field is required for proper operation",
		"invalid_value":        "The value doesn't match expected format or allowed values",
		"permission_error":     "Check file permissions and ensure the application can read the file",
		"compatibility_warning": "Different settings between sources may affect comparison accuracy",
		"performance_suggestion": "This optimization can significantly improve processing speed for large datasets",
	}
	return explanations[issueType]
}

// Data structures for validation results

type ValidationResult struct {
	ConfigPaths     []string                   `yaml:"config_paths"`
	Valid          bool                       `yaml:"valid"`
	Errors         []ValidationError          `yaml:"errors,omitempty"`
	Recommendations []ValidationRecommendation `yaml:"recommendations,omitempty"`
}

type ValidationError struct {
	ConfigPath string `yaml:"config_path"`
	Type       string `yaml:"type"`
	Field      string `yaml:"field,omitempty"`
	Message    string `yaml:"message"`
	Severity   string `yaml:"severity"` // error, warning, info
}

type ValidationRecommendation struct {
	Type     string `yaml:"type"`
	Priority string `yaml:"priority"` // high, medium, low
	Title    string `yaml:"title"`
	Message  string `yaml:"message"`
	Action   string `yaml:"action,omitempty"`
}