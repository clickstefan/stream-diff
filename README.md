# Data Stream Comparator

A powerful command-line tool written in Go to compare two arbitrary streams of data, identify differences, and report detailed statistics.

## Overview

This tool is designed to address the challenge of comparing large datasets from various sources, such as CSV files or Kafka topics. It performs a deep comparison by first learning the schema of the data, including data types and patterns, and then provides a comprehensive report on how two data sources differ.

This is particularly useful for data validation, migration testing, and ensuring data integrity between different systems.

## Features

- **Multiple Data Sources:** Supports reading from different sources, starting with CSV and JSON-Lines (`.jsonl`).
- **Automatic Schema Detection:**
    - Infers the schema from a sample of the data.
    - Flattens nested JSON objects and arrays into a dot-notation format (e.g., `customer.address.city`).
    - Automatically detects data types: `numeric`, `string`, `boolean`, `date`, `datetime`, `timestamp`.
- **Advanced String Parsing:**
    - Can detect and recursively parse JSON strings embedded within other file formats (e.g., a CSV field containing a JSON object).
    - Identifies field patterns using a library of built-in regex matchers and supports custom matchers.
    - **NEW**: Optional AI-powered pattern detection using embedded models (offline) or Claude/Anthropic APIs (online) to automatically generate regex patterns for data validation.
- **Intelligent Date/Time Handling:**
    - Parses and compares `date`, `datetime`, and `timestamp` fields, even if their string formats differ between sources.
    - Supports timestamps with variable precision.
- **Comprehensive Reporting:**
    - Generates a detailed comparison report in YAML format.
    - **Summary:** High-level overview of the comparison (rows processed, matching keys, etc.).
    - **Value Diffs:** A list of records that have the same key but different values in other fields.
    - **Keys Only:** Lists of keys found only in one source.
    - **Field Stats:** A complete statistical profile for every field in both data sources, including `min`, `max`, `avg`, `cardinality`, `availability`, and `avgDaysAgo` for date/time fields.

## Configuration

The tool is configured using two YAML files, one for each data source.

**Example `config.yaml`:**
```yaml
source:
  # Type of the data source. Supported: csv, json
  type: csv
  # Path to the source file.
  path: path/to/your/data.csv
  # Optional parser configuration.
  parser_config:
    # Set to true to enable recursive parsing of string fields that look like JSON.
    json_in_string: true

# Optional: Enable AI-powered pattern detection
pattern_detection:
  enabled: true
  mode: offline  # or "online" for Claude/Anthropic API
  
  # Offline mode (built-in pattern recognition)
  offline_model:
    # Uses built-in patterns for common data types
    
  # Online mode configuration (for Claude/Anthropic API)
  # online_api:
  #   provider: claude  # or "anthropic"
  #   api_key: "your-api-key-here"
  #   model: "claude-3-haiku-20240307"  # optional

# Optional: Define a schema to use instead of generating one.
# schema:
#   key: user_id
#   fields:
#     ...
```

## Usage

To run a comparison, use the `compare` command and provide the paths to the two configuration files.

```bash
# (Once implemented)
go run ./cmd/comparator compare ./config1.yaml ./config2.yaml
```

## AI-Powered Pattern Detection

This tool includes optional AI-powered pattern detection to automatically identify regex patterns in your data fields and enhance schema generation. This feature helps complete the schema with appropriate matchers for data validation.

### Offline Mode
Uses built-in pattern recognition for common data types:
- Email addresses
- Phone numbers  
- URLs
- IP addresses
- UUIDs
- Numeric values
- Date/time values

### Online Mode
Integrates with AI services (Claude/Anthropic) for more sophisticated pattern detection:
- Analyzes field samples using AI
- Generates custom regex patterns
- Supports complex data patterns beyond built-in types

### Configuration
Enable pattern detection in your config file:
```yaml
pattern_detection:
  enabled: true
  mode: offline  # or "online"
  
  # For online mode:
  online_api:
    provider: claude
    api_key: "your-api-key"
    model: "claude-3-haiku-20240307"
```

## Testing

This project is developed using a test-driven approach. A comprehensive suite of test cases, including source data and expected outputs, can be found in the `testdata` directory. These tests cover all major features and edge cases.
