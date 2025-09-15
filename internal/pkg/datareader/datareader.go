package datareader

import (
	"data-comparator/internal/pkg/config"
	"data-comparator/internal/pkg/generator"
	"data-comparator/internal/pkg/types"
	"fmt"
)

// Record represents a single record from a data source, like a CSV row or a JSON object.
type Record = types.Record

// DataReader is the interface for reading records from a data source.
type DataReader = types.DataReader

// New creates a new DataReader based on the provided source configuration.
func New(cfg config.Source) (DataReader, error) {
	switch cfg.Type {
	case "csv":
		return NewCSVReader(cfg)
	case "json":
		return NewJSONReader(cfg)
	case "stream":
		return generator.NewFromConfig(cfg)
	default:
		return nil, fmt.Errorf("unsupported source type: %s", cfg.Type)
	}
}
