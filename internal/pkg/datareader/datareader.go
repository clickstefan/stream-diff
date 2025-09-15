package datareader

import (
	"data-comparator/internal/pkg/config"
	"fmt"
)

// Record represents a single record from a data source, like a CSV row or a JSON object.
type Record map[string]interface{}

// DataReader is the interface for reading records from a data source.
type DataReader interface {
	// Read returns the next record from the source.
	// It returns io.EOF when there are no more records.
	Read() (Record, error)
	// Close closes the reader and any underlying resources.
	Close() error
}

// New creates a new DataReader based on the provided source configuration.
func New(cfg config.Source) (DataReader, error) {
	switch cfg.Type {
	case "csv":
		return NewCSVReader(cfg)
	case "json":
		return NewJSONReader(cfg)
	case "protobuf", "proto":
		return NewProtobufReader(cfg)
	default:
		return nil, fmt.Errorf("unsupported source type: %s", cfg.Type)
	}
}
