package datareader

import (
	"data-comparator/internal/pkg/config"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// CSVReader reads records from a CSV file.
type CSVReader struct {
	file         *os.File
	reader       *csv.Reader
	header       []string
	parserConfig config.ParserConfig
}

// NewCSVReader creates a new reader for CSV files.
func NewCSVReader(cfg config.Source) (DataReader, error) {
	file, err := os.Open(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open csv file %s: %w", cfg.Path, err)
	}

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("csv file %s is empty", cfg.Path)
		}
		return nil, fmt.Errorf("failed to read header from csv file %s: %w", cfg.Path, err)
	}

	var pcfg config.ParserConfig
	if cfg.ParserConfig != nil {
		pcfg = *cfg.ParserConfig
	}

	return &CSVReader{
		file:         file,
		reader:       reader,
		header:       header,
		parserConfig: pcfg,
	}, nil
}

// Read reads the next record from the CSV file.
func (r *CSVReader) Read() (Record, error) {
	row, err := r.reader.Read()
	if err != nil {
		return nil, err // This will correctly return io.EOF at the end of the file
	}

	record := make(Record)
	for i, value := range row {
		if i < len(r.header) {
			var processedValue interface{} = value
			if r.parserConfig.JSONInString {
				processedValue = r.tryParseJSON(value)
			}
			record[r.header[i]] = processedValue
		}
	}
	return record, nil
}

// tryParseJSON attempts to recursively unmarshal a string as JSON.
// If it fails, it returns the original string.
func (r *CSVReader) tryParseJSON(s string) interface{} {
	if s == "" {
		return s
	}

	var result interface{}
	err := json.Unmarshal([]byte(s), &result)
	if err != nil {
		return s
	}

	if strVal, ok := result.(string); ok {
		return r.tryParseJSON(strVal)
	}

	if mapVal, ok := result.(map[string]interface{}); ok {
		for k, v := range mapVal {
			if strV, ok := v.(string); ok {
				mapVal[k] = r.tryParseJSON(strV)
			}
		}
		return mapVal
	}

	return result
}

// Close closes the underlying file.
func (r *CSVReader) Close() error {
	return r.file.Close()
}
