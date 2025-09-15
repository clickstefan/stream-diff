package datareader

import (
	"data-comparator/internal/pkg/config"
	"fmt"
	"os"
	"bufio"
	"io"
	"encoding/json"
)

// ProtobufReader reads records from protobuf files.
// It supports different protobuf formats:
// - Binary protobuf messages (requires descriptor file)
// - JSON serialized protobuf messages 
// - Text format protobuf messages
type ProtobufReader struct {
	file         *os.File
	scanner      *bufio.Scanner
	format       string // "binary", "json", "text"
	parserConfig config.ParserConfig
}

// ProtobufParserConfig extends ParserConfig with protobuf-specific options
type ProtobufParserConfig struct {
	config.ParserConfig
	// Format specifies the protobuf format: "binary", "json", or "text"
	// Default is "json" which is most common for streaming data
	Format string `yaml:"format"`
	// DescriptorPath is the path to the protobuf descriptor file (.desc)
	// Required for binary format, optional for others
	DescriptorPath string `yaml:"descriptor_path"`
	// MessageType is the name of the protobuf message type
	// Required when using descriptor file
	MessageType string `yaml:"message_type"`
	// MessageSeparator for binary format (default is length-prefixed)
	// Can be "length-prefixed", "newline", or "fixed-size"
	MessageSeparator string `yaml:"message_separator"`
}

// NewProtobufReader creates a new reader for protobuf files.
func NewProtobufReader(cfg config.Source) (DataReader, error) {
	file, err := os.Open(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open protobuf file %s: %w", cfg.Path, err)
	}

	// Default to JSON format for protobuf
	format := "json"

	// Parse protobuf-specific config if provided
	if cfg.ParserConfig != nil {
		// For now, we'll detect format from the config or file extension
		// This is a simplified approach - in a real implementation,
		// we'd want a more robust protobuf configuration structure
		format = "json" // Default assumption for streaming protobuf data
	}

	var pcfg config.ParserConfig
	if cfg.ParserConfig != nil {
		pcfg = *cfg.ParserConfig
	}

	reader := &ProtobufReader{
		file:         file,
		format:       format,
		parserConfig: pcfg,
	}

	// Initialize scanner for text-based formats
	if format == "json" || format == "text" {
		reader.scanner = bufio.NewScanner(file)
	}

	return reader, nil
}

// Read reads the next record from the protobuf file.
func (r *ProtobufReader) Read() (Record, error) {
	switch r.format {
	case "json":
		return r.readJSONFormat()
	case "text":
		return r.readTextFormat()
	case "binary":
		return r.readBinaryFormat()
	default:
		return nil, fmt.Errorf("unsupported protobuf format: %s", r.format)
	}
}

// readJSONFormat reads JSON-serialized protobuf messages (most common for streaming)
func (r *ProtobufReader) readJSONFormat() (Record, error) {
	if !r.scanner.Scan() {
		err := r.scanner.Err()
		if err != nil {
			return nil, err
		}
		return nil, io.EOF
	}

	line := r.scanner.Text()
	if line == "" {
		return r.Read() // Skip empty lines
	}

	var record Record
	err := json.Unmarshal([]byte(line), &record)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON protobuf message: %w", err)
	}

	// Apply recursive JSON parsing if enabled
	if r.parserConfig.JSONInString {
		record = r.processJSONInString(record)
	}

	return record, nil
}

// readTextFormat reads text format protobuf messages
func (r *ProtobufReader) readTextFormat() (Record, error) {
	if !r.scanner.Scan() {
		err := r.scanner.Err()
		if err != nil {
			return nil, err
		}
		return nil, io.EOF
	}

	line := r.scanner.Text()
	if line == "" {
		return r.Read() // Skip empty lines
	}

	// For text format, we'll convert to JSON first for easier processing
	// This is a simplified approach - real text format parsing would be more complex
	record := make(Record)
	record["raw_text"] = line
	// TODO: Implement proper text format parsing when needed
	
	return record, nil
}

// readBinaryFormat reads binary protobuf messages
func (r *ProtobufReader) readBinaryFormat() (Record, error) {
	// For binary format, we would need the message descriptor
	// This would need to be implemented when binary protobuf support is required
	return nil, fmt.Errorf("binary protobuf format not yet implemented")
}

// processJSONInString applies recursive JSON parsing to string fields
func (r *ProtobufReader) processJSONInString(data Record) Record {
	result := make(Record)
	for k, v := range data {
		switch val := v.(type) {
		case string:
			result[k] = r.tryParseJSON(val)
		case map[string]interface{}:
			result[k] = r.processJSONInString(Record(val))
		case []interface{}:
			result[k] = r.processArray(val)
		default:
			result[k] = val
		}
	}
	return result
}

// processArray applies recursive JSON parsing to array elements
func (r *ProtobufReader) processArray(arr []interface{}) []interface{} {
	result := make([]interface{}, len(arr))
	for i, v := range arr {
		switch val := v.(type) {
		case string:
			result[i] = r.tryParseJSON(val)
		case map[string]interface{}:
			result[i] = r.processJSONInString(Record(val))
		case []interface{}:
			result[i] = r.processArray(val)
		default:
			result[i] = val
		}
	}
	return result
}

// tryParseJSON attempts to recursively unmarshal a string as JSON.
// This is similar to the CSV reader's implementation.
func (r *ProtobufReader) tryParseJSON(s string) interface{} {
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
func (r *ProtobufReader) Close() error {
	return r.file.Close()
}