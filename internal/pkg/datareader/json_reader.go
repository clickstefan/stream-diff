package datareader

import (
	"data-comparator/internal/pkg/config"
	"encoding/json"
	"fmt"
	"os"
)

// JSONReader reads records from a JSON-Lines file.
type JSONReader struct {
	file    *os.File
	decoder *json.Decoder
}

// NewJSONReader creates a new reader for JSON-Lines files.
func NewJSONReader(cfg config.Source) (DataReader, error) {
	file, err := os.Open(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open json file %s: %w", cfg.Path, err)
	}

	return &JSONReader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

// Read reads the next record from the JSON-Lines file.
func (r *JSONReader) Read() (Record, error) {
	var record Record
	err := r.decoder.Decode(&record) // Decode will return io.EOF at the end.
	if err != nil {
		return nil, err
	}
	return record, nil
}

// Close closes the underlying file.
func (r *JSONReader) Close() error {
	return r.file.Close()
}
