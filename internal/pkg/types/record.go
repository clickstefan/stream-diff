package types

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