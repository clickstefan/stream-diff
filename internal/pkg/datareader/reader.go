package datareader

// Reader interface for reading data
type Reader interface {
	Read() ([]byte, error)
}

// CSVReader reads CSV data
type CSVReader struct {
	// Add fields as needed
}

// JSONReader reads JSON data
type JSONReader struct {
	// Add fields as needed
}