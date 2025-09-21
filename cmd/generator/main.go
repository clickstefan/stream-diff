package main

import (
	"data-comparator/internal/pkg/config"
	"data-comparator/internal/pkg/datareader"
	"data-comparator/internal/pkg/types"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// OutputFormat defines the supported output formats
type OutputFormat string

const (
	FormatCSV    OutputFormat = "csv"
	FormatJSONL  OutputFormat = "jsonl"
	FormatProto  OutputFormat = "proto"
)

type Config struct {
	SchemaPath       string
	Format           OutputFormat
	MaxRecords       int64
	RecordsPerSecond int64
	BufferSize       int
	Seed             int64
	Delimiter        string
	Header           bool
}

func main() {
	cfg := parseFlags()
	
	// Create stream generator config
	streamConfig := config.StreamGeneratorConfig{
		SchemaPath:       cfg.SchemaPath,
		MaxRecords:       cfg.MaxRecords,
		RecordsPerSecond: float64(cfg.RecordsPerSecond),
		BufferSize:       cfg.BufferSize,
		Seed:             cfg.Seed,
		// Add realistic data patterns for various real-world scenarios
		DataPatterns:     getRealWorldDataPatterns(),
	}
	
	source := config.Source{
		Type:            "stream",
		StreamGenerator: &streamConfig,
	}
	
	// Create the data reader (generator)
	reader, err := datareader.New(source)
	if err != nil {
		log.Fatalf("Failed to create stream generator: %v", err)
	}
	defer reader.Close()
	
	// Output data in the specified format
	err = outputData(reader, cfg)
	if err != nil {
		log.Fatalf("Failed to output data: %v", err)
	}
}

func parseFlags() *Config {
	cfg := &Config{}
	
	var format string
	flag.StringVar(&cfg.SchemaPath, "schema", "", "Path to schema YAML file (optional)")
	flag.StringVar(&format, "format", "jsonl", "Output format: csv, jsonl, proto")
	flag.Int64Var(&cfg.MaxRecords, "count", 100, "Maximum number of records to generate (0 = unlimited)")
	flag.Int64Var(&cfg.RecordsPerSecond, "rate", 0, "Records per second (0 = unlimited)")
	flag.IntVar(&cfg.BufferSize, "buffer", 100, "Buffer size for backpressure handling")
	flag.Int64Var(&cfg.Seed, "seed", 0, "Random seed for reproducible output (0 = use current time)")
	flag.StringVar(&cfg.Delimiter, "delimiter", "", "Custom delimiter (default: \\n for csv/jsonl, \\0 for proto)")
	flag.BoolVar(&cfg.Header, "header", true, "Include CSV header row (CSV format only)")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nGenerates realistic test data for databases, Kafka topics, logs, etc.\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s -format csv -count 1000 -header > data.csv\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -format jsonl -rate 100 -count 5000 | kafka-console-producer.sh\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -schema user_schema.yaml -format proto -count 1000\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}
	
	flag.Parse()
	
	// Validate and set format
	switch strings.ToLower(format) {
	case "csv":
		cfg.Format = FormatCSV
	case "jsonl", "json":
		cfg.Format = FormatJSONL
	case "proto", "protobuf":
		cfg.Format = FormatProto
	default:
		log.Fatalf("Unsupported format: %s. Supported formats: csv, jsonl, proto", format)
	}
	
	// Set default delimiters if not specified
	if cfg.Delimiter == "" {
		switch cfg.Format {
		case FormatCSV, FormatJSONL:
			cfg.Delimiter = "\n"
		case FormatProto:
			cfg.Delimiter = "\n" // Use newline for proto as well for easier piping
		}
	}
	
	return cfg
}

func outputData(reader types.DataReader, cfg *Config) error {
	switch cfg.Format {
	case FormatCSV:
		return outputCSV(reader, cfg)
	case FormatJSONL:
		return outputJSONL(reader, cfg)
	case FormatProto:
		return outputProto(reader, cfg)
	default:
		return fmt.Errorf("unsupported output format: %s", cfg.Format)
	}
}

func outputCSV(reader types.DataReader, cfg *Config) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()
	
	var headers []string
	headerWritten := false
	recordCount := int64(0)
	
	for {
		if cfg.MaxRecords > 0 && recordCount >= cfg.MaxRecords {
			break
		}
		
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading record: %v", err)
		}
		
		// Extract headers from first record
		if !headerWritten {
			headers = extractHeaders(record)
			if cfg.Header {
				if err := writer.Write(headers); err != nil {
					return fmt.Errorf("error writing CSV header: %v", err)
				}
			}
			headerWritten = true
		}
		
		// Convert record to CSV row
		row := recordToCSVRow(record, headers)
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing CSV row: %v", err)
		}
		
		recordCount++
		
		// Custom delimiter handling (flush and add delimiter)
		if cfg.Delimiter != "\n" {
			writer.Flush()
			fmt.Print(cfg.Delimiter)
		}
	}
	
	return nil
}

func outputJSONL(reader types.DataReader, cfg *Config) error {
	encoder := json.NewEncoder(os.Stdout)
	recordCount := int64(0)
	
	for {
		if cfg.MaxRecords > 0 && recordCount >= cfg.MaxRecords {
			break
		}
		
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading record: %v", err)
		}
		
		if err := encoder.Encode(record); err != nil {
			return fmt.Errorf("error encoding JSON record: %v", err)
		}
		
		recordCount++
		
		// Custom delimiter handling
		if cfg.Delimiter != "\n" {
			fmt.Print(cfg.Delimiter)
		}
	}
	
	return nil
}

func outputProto(reader types.DataReader, cfg *Config) error {
	// For now, output as JSON with protobuf-like structure
	// TODO: Add actual protobuf serialization when schema definitions are available
	recordCount := int64(0)
	
	for {
		if cfg.MaxRecords > 0 && recordCount >= cfg.MaxRecords {
			break
		}
		
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading record: %v", err)
		}
		
		// Convert to protobuf-style JSON
		protoJSON, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("error marshaling record to protobuf JSON: %v", err)
		}
		
		fmt.Print(string(protoJSON))
		fmt.Print(cfg.Delimiter)
		
		recordCount++
	}
	
	return nil
}

// Helper functions

func extractHeaders(record types.Record) []string {
	var headers []string
	for key := range record {
		headers = append(headers, key)
	}
	return headers
}

func recordToCSVRow(record types.Record, headers []string) []string {
	row := make([]string, len(headers))
	for i, header := range headers {
		value := record[header]
		row[i] = valueToString(value)
	}
	return row
}

func valueToString(value interface{}) string {
	if value == nil {
		return ""
	}
	
	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%.2f", v)
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		// For complex types, serialize as JSON
		jsonBytes, _ := json.Marshal(v)
		return string(jsonBytes)
	}
}

// getRealWorldDataPatterns returns comprehensive data patterns for various real-world scenarios
func getRealWorldDataPatterns() map[string]config.DataPattern {
	return map[string]config.DataPattern{
		// E-commerce patterns
		"product_id": {Type: "format", Format: "PROD-{id}"},
		"order_id": {Type: "format", Format: "ORD-{id}"},
		"sku": {Type: "format", Format: "SKU-{random}"},
		"price": {Type: "range", Min: 9.99, Max: 999.99},
		"category": {
			Type: "list",
			Values: []interface{}{"electronics", "clothing", "books", "home-garden", "sports", "toys", "automotive"},
		},
		"payment_method": {
			Type: "list",
			Values: []interface{}{"credit_card", "debit_card", "paypal", "bank_transfer", "cash", "crypto"},
		},
		"order_status": {
			Type: "list",
			Values: []interface{}{"pending", "confirmed", "processing", "shipped", "delivered", "cancelled", "refunded"},
		},
		
		// User/Customer patterns
		"user_id": {Type: "format", Format: "user_{id}"},
		"customer_id": {Type: "format", Format: "CUST-{id}"},
		"email": {Type: "format", Format: "email"},
		"phone": {Type: "format", Format: "phone"},
		"age": {Type: "range", Min: 18, Max: 85},
		"gender": {
			Type: "list",
			Values: []interface{}{"male", "female", "non-binary", "prefer-not-to-say"},
		},
		"plan_type": {
			Type: "list",
			Values: []interface{}{"free", "basic", "premium", "enterprise", "trial"},
		},
		"subscription_status": {
			Type: "list",
			Values: []interface{}{"active", "inactive", "cancelled", "expired", "pending"},
		},
		
		// Geographic patterns
		"country": {
			Type: "list",
			Values: []interface{}{"USA", "Canada", "UK", "Germany", "France", "Australia", "Japan", "Brazil", "India", "Mexico", "China", "Russia"},
		},
		"city": {
			Type: "list",
			Values: []interface{}{"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia", "San Antonio", "San Diego", "Dallas", "San Jose", "Austin", "Jacksonville", "Fort Worth", "Columbus", "Charlotte", "San Francisco", "Indianapolis", "Seattle", "Denver", "Washington DC", "Boston", "El Paso", "Nashville", "Detroit", "Oklahoma City"},
		},
		"timezone": {
			Type: "list",
			Values: []interface{}{"UTC", "America/New_York", "America/Chicago", "America/Denver", "America/Los_Angeles", "Europe/London", "Europe/Paris", "Asia/Tokyo", "Asia/Shanghai", "Australia/Sydney"},
		},
		
		// Technical/System patterns
		"ip_address": {Type: "format", Format: "ip"},
		"mac_address": {Type: "format", Format: "mac"},
		"uuid": {Type: "format", Format: "uuid"},
		"session_id": {Type: "format", Format: "uuid"},
		"api_key": {Type: "format", Format: "api_key"},
		"version": {
			Type: "list",
			Values: []interface{}{"v1.0.0", "v1.1.0", "v1.2.0", "v2.0.0", "v2.1.0", "v3.0.0"},
		},
		"browser": {
			Type: "list",
			Values: []interface{}{"Chrome", "Firefox", "Safari", "Edge", "Opera", "Mobile Safari", "Chrome Mobile"},
		},
		"os": {
			Type: "list",
			Values: []interface{}{"Windows", "macOS", "Linux", "iOS", "Android", "Ubuntu", "CentOS"},
		},
		"device_type": {
			Type: "list",
			Values: []interface{}{"desktop", "mobile", "tablet", "smart-tv", "wearable", "iot"},
		},
		
		// Log/Event patterns
		"log_level": {
			Type: "list",
			Values: []interface{}{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"},
		},
		"event_type": {
			Type: "list",
			Values: []interface{}{"user_login", "user_logout", "page_view", "click", "purchase", "search", "error", "api_call"},
		},
		"http_status": {
			Type: "list",
			Values: []interface{}{200, 201, 400, 401, 403, 404, 500, 502, 503},
		},
		"response_time": {Type: "range", Min: 10, Max: 5000},
		
		// Financial patterns
		"transaction_id": {Type: "format", Format: "TXN-{id}"},
		"account_number": {Type: "format", Format: "ACC-{id}"},
		"amount": {Type: "range", Min: 1.00, Max: 10000.00},
		"currency": {
			Type: "list",
			Values: []interface{}{"USD", "EUR", "GBP", "JPY", "CAD", "AUD", "CNY", "INR", "BRL", "MXN"},
		},
		"transaction_type": {
			Type: "list",
			Values: []interface{}{"debit", "credit", "transfer", "payment", "refund", "fee", "interest"},
		},
		
		// IoT/Sensor patterns
		"sensor_id": {Type: "format", Format: "SENSOR-{id}"},
		"temperature": {Type: "range", Min: -20.0, Max: 45.0},
		"humidity": {Type: "range", Min: 0.0, Max: 100.0},
		"pressure": {Type: "range", Min: 950.0, Max: 1050.0},
		"battery_level": {Type: "range", Min: 0, Max: 100},
		"signal_strength": {Type: "range", Min: -100, Max: -30},
		
		// Gaming patterns
		"player_id": {Type: "format", Format: "PLAYER-{id}"},
		"score": {Type: "range", Min: 0, Max: 999999},
		"level": {Type: "range", Min: 1, Max: 100},
		"game_mode": {
			Type: "list",
			Values: []interface{}{"single_player", "multiplayer", "coop", "tournament", "practice"},
		},
		
		// Media patterns
		"media_id": {Type: "format", Format: "MEDIA-{id}"},
		"duration": {Type: "range", Min: 30, Max: 7200}, // seconds
		"quality": {
			Type: "list",
			Values: []interface{}{"240p", "360p", "480p", "720p", "1080p", "1440p", "4K"},
		},
		"content_type": {
			Type: "list",
			Values: []interface{}{"video", "audio", "image", "document", "stream"},
		},
	}
}