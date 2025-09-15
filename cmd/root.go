package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
	debug   bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "stream-diff",
	Short: "A powerful tool to compare data streams and identify differences",
	Long: `Stream-Diff is a comprehensive command-line tool for comparing data streams
from various sources such as CSV files, JSON files, and more.

It performs intelligent schema detection, data type inference, and provides
detailed comparison reports with statistics and insights.

Features:
  • Multiple data source support (CSV, JSON-Lines)
  • Automatic schema detection and type inference
  • Advanced string parsing with embedded JSON support
  • Intelligent date/time handling with format flexibility
  • Comprehensive statistical reporting
  • AI-powered insights and recommendations`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initLogging()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.stream-diff.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get user home directory")
		}

		// Search config in home directory with name ".stream-diff" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".stream-diff")
	}

	// Environment variables
	viper.SetEnvPrefix("STREAM_DIFF")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Info().Str("config", viper.ConfigFileUsed()).Msg("Using config file")
	}
}

// initLogging configures the logging system based on flags
func initLogging() {
	// Configure zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Set log level
	logLevel := zerolog.InfoLevel
	if debug || viper.GetBool("debug") {
		logLevel = zerolog.DebugLevel
	} else if verbose || viper.GetBool("verbose") {
		logLevel = zerolog.InfoLevel
	} else {
		logLevel = zerolog.WarnLevel
	}

	zerolog.SetGlobalLevel(logLevel)

	// Configure console output for better UX
	if isTerminal() {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "15:04:05",
		})
	}
}

// isTerminal checks if we're running in a terminal
func isTerminal() bool {
	fileInfo, _ := os.Stderr.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// AI-powered help system
func aiEnhancedHelp(cmd *cobra.Command) string {
	baseHelp := cmd.Long
	if baseHelp == "" {
		baseHelp = cmd.Short
	}

	// Add contextual AI suggestions based on command
	suggestions := getAISuggestions(cmd.Use)
	if suggestions != "" {
		return fmt.Sprintf("%s\n\n🤖 AI Suggestions:\n%s", baseHelp, suggestions)
	}

	return baseHelp
}

// getAISuggestions provides context-aware suggestions
func getAISuggestions(cmdName string) string {
	switch cmdName {
	case "compare":
		return `• Start with small datasets to understand the output format
• Use --sample-size to limit processing for large files
• Enable --json-in-string for CSV files containing JSON data
• Check the schema output first with --schema-only flag`
	case "validate":
		return `• Validation helps catch configuration issues early
• Use --explain for detailed validation reports
• Check file paths and permissions before running comparisons`
	default:
		return `• Use 'stream-diff compare --help' to see comparison options
• Run 'stream-diff validate' to check your configuration files
• Enable verbose mode (-v) for detailed progress information`
	}
}