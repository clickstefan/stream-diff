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
# Optional: Define a schema to use instead of generating one.
# schema:
#   key: user_id
#   fields:
#     ...
```

## Usage

### Command Line Interface

The tool can be run in two ways:

#### 1. Using a Run Configuration File (Recommended)

```bash
go run ./cmd/stream-diff -config runConfig.yaml -key user_id
```

#### 2. Using Command Line Parameters

```bash
go run ./cmd/stream-diff \
  -source1 path/to/source1.csv \
  -source2 path/to/source2.csv \
  -key user_id \
  -enable-periodic \
  -time-interval 30 \
  -record-interval 1000 \
  -output-dir ./reports
```

### Run Configuration File

Create a `runConfig.yaml` file to define your comparison settings:

```yaml
source1:
  type: csv
  path: data/source1.csv
  parser_config:
    json_in_string: false

source2:
  type: csv  
  path: data/source2.csv
  parser_config:
    json_in_string: false

output:
  final_report: final_report.yaml
  periodic_reports: periodic_reports

periodic:
  enabled: true
  time_interval_seconds: 30
  record_interval: 1000
```

### Periodic Diff Reporting

The tool supports periodic reporting of differences as data streams are processed. This is useful for:
- Monitoring long-running comparisons
- Early detection of data differences
- Progress tracking for large datasets

**Configuration Options:**
- `time_interval_seconds`: Generate reports every N seconds
- `record_interval`: Generate reports every N records processed
- Both options can be used together - reports trigger when either condition is met

**Output:**
- Periodic reports are saved to timestamped YAML files in the specified directory
- Console output shows real-time progress updates
- Final comprehensive report is generated at the end

## Testing

This project is developed using a test-driven approach. A comprehensive suite of test cases, including source data and expected outputs, can be found in the `testdata` directory. These tests cover all major features and edge cases.
