# Data Stream Comparator

A powerful command-line tool written in Go to compare two arbitrary streams of data, identify differences, and report detailed statistics.

## Overview

This tool is designed to address the challenge of comparing large datasets from various sources, such as CSV files or Kafka topics. It performs a deep comparison by first learning the schema of the data, including data types and patterns, and then provides a comprehensive report on how two data sources differ.

This is particularly useful for data validation, migration testing, and ensuring data integrity between different systems.

## Features

- **Multiple Data Sources:** Supports reading from different sources, including CSV, JSON-Lines (`.jsonl`), and **stream generators** for performance testing.
- **Stream Generator:** Generate realistic test data based on schemas with configurable patterns, rate limiting, and backpressure handling.
- **Automatic Schema Detection:**
    - Infers the schema from a sample of the data.
    - Flattens nested JSON objects and arrays into a dot-notation format (e.g., `customer.address.city`).
    - Automatically detects data types: `numeric`, `string`, `boolean`, `date`, `datetime`, `timestamp`.
- **Advanced String Parsing:**
    - Can detect and recursively parse JSON strings embedded within other file formats (e.g., a CSV field containing a JSON object).
    - Identifies field patterns using a library of built-in regex matchers and supports custom matchers.
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
  # Type of the data source. Supported: csv, json, stream
  type: csv
  # Path to the source file.
  path: path/to/your/data.csv
  # Optional parser configuration.
  parser_config:
    # Set to true to enable recursive parsing of string fields that look like JSON.
    json_in_string: true
# Optional: Define a schema to use instead of generating one.
# schema:
#   key: user_id
#   fields:
#     ...
```

**Stream Generator Configuration:**
```yaml
source:
  type: stream
  stream_generator:
    # Path to schema file that defines the structure of generated data
    schema_path: examples/user_schema.yaml
    
    # Generate 10,000 records (0 = unlimited)
    max_records: 10000
    
    # Generate 100 records per second (0 = no rate limiting)
    records_per_second: 100
    
    # Buffer size for backpressure handling
    buffer_size: 500
    
    # Random seed for reproducible data generation (0 = use current time)
    seed: 42
    
    # Custom data patterns for specific fields
    data_patterns:
      plan_type:
        type: list
        values: ["basic", "premium", "enterprise"]
      
      age:
        type: range
        min: 18
        max: 85
      
      email:
        type: format
        format: email
```

## Usage

To run a comparison, use the `compare` command and provide the paths to the two configuration files.

```bash
# (Once implemented)
go run ./cmd/comparator compare ./config1.yaml ./config2.yaml
```

### Stream Generator Demo

To test the stream generator functionality, you can use the provided demo:

```bash
# Generate and display 10 sample records using the example configuration
go run examples/stream_demo.go examples/stream_config.yaml 10
```

The stream generator provides:

- **Realistic Data Generation:** Generates data based on field names and types (e.g., emails, names, dates)
- **Custom Patterns:** Define custom value lists, ranges, or formats for specific fields
- **Rate Limiting:** Control generation speed to simulate real-world data streams
- **Backpressure Handling:** Uses buffered channels to prevent memory overflow when readers are slow
- **Reproducible Output:** Use seeds for consistent data generation across runs

## Testing

This project is developed using a test-driven approach. A comprehensive suite of test cases, including source data and expected outputs, can be found in the `testdata` directory. These tests cover all major features and edge cases.
