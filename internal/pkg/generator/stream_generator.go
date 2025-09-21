package generator

import (
	"data-comparator/internal/pkg/config"
	"data-comparator/internal/pkg/schema"
	"data-comparator/internal/pkg/types"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// StreamGenerator generates realistic test data based on a schema.
// It implements DataReader interface and handles backpressure.
type StreamGenerator struct {
	schema       *schema.Schema
	config       config.StreamGeneratorConfig
	recordCh     chan types.Record
	stopCh       chan struct{}
	wg           sync.WaitGroup
	mu           sync.RWMutex
	closed       bool
	recordCount  int64
	maxRecords   int64
	rng          *rand.Rand
}

// NewFromConfig creates a stream generator from configuration.
func NewFromConfig(cfg config.Source) (types.DataReader, error) {
	if cfg.StreamGenerator == nil {
		return nil, fmt.Errorf("stream_generator configuration is required for stream type")
	}
	
	// Load schema from file if specified
	var schemaObj *schema.Schema
	if cfg.StreamGenerator.SchemaPath != "" {
		data, err := os.ReadFile(cfg.StreamGenerator.SchemaPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read schema file %s: %w", cfg.StreamGenerator.SchemaPath, err)
		}
		
		schemaObj = &schema.Schema{}
		err = yaml.Unmarshal(data, schemaObj)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal schema from %s: %w", cfg.StreamGenerator.SchemaPath, err)
		}
	} else {
		// Use a default schema if none provided
		schemaObj = createDefaultSchema()
	}
	
	return New(schemaObj, *cfg.StreamGenerator)
}

// createDefaultSchema creates a basic schema for testing purposes.
func createDefaultSchema() *schema.Schema {
	return &schema.Schema{
		Key: "user_id",
		Fields: map[string]*schema.Field{
			"user_id":    {Type: "numeric"},
			"email":      {Type: "string"},
			"age":        {Type: "numeric"},
			"city":       {Type: "string"},
			"plan_type":  {Type: "string"},
			"last_login": {Type: "datetime"},
			"active":     {Type: "boolean"},
		},
	}
}

// New creates a new stream generator.
func New(schema *schema.Schema, config config.StreamGeneratorConfig) (*StreamGenerator, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema is required")
	}
	
	// Set defaults
	if config.BufferSize <= 0 {
		config.BufferSize = 100
	}
	if config.Seed == 0 {
		config.Seed = time.Now().UnixNano()
	}
	
	gen := &StreamGenerator{
		schema:      schema,
		config:      config,
		recordCh:    make(chan types.Record, config.BufferSize),
		stopCh:      make(chan struct{}),
		maxRecords:  config.MaxRecords,
		rng:         rand.New(rand.NewSource(config.Seed)),
	}
	
	// Start the generator goroutine
	gen.wg.Add(1)
	go gen.generateRecords()
	
	return gen, nil
}

// Read returns the next generated record.
// It returns io.EOF when the generator is finished or closed.
func (g *StreamGenerator) Read() (types.Record, error) {
	select {
	case record, ok := <-g.recordCh:
		if !ok {
			return nil, io.EOF
		}
		return record, nil
	case <-g.stopCh:
		return nil, io.EOF
	}
}

// Close stops the generator and cleans up resources.
func (g *StreamGenerator) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	if g.closed {
		return nil
	}
	
	g.closed = true
	close(g.stopCh)
	g.wg.Wait()
	
	// Drain the channel
	for range g.recordCh {
		// Drain remaining records
	}
	
	return nil
}

// generateRecords runs in a goroutine and generates records according to the schema and config.
func (g *StreamGenerator) generateRecords() {
	defer g.wg.Done()
	defer close(g.recordCh)
	
	var ticker *time.Ticker
	var tickerCh <-chan time.Time
	
	// Set up rate limiting if configured
	if g.config.RecordsPerSecond > 0 {
		interval := time.Duration(float64(time.Second) / g.config.RecordsPerSecond)
		ticker = time.NewTicker(interval)
		defer ticker.Stop()
		tickerCh = ticker.C
	}
	
	recordID := int64(1)
	
	for {
		// Check if we've reached the maximum record count
		if g.maxRecords > 0 && recordID > g.maxRecords {
			return
		}
		
		// Check for stop signal
		select {
		case <-g.stopCh:
			return
		default:
		}
		
		// Rate limiting
		if tickerCh != nil {
			select {
			case <-tickerCh:
				// Continue to generate
			case <-g.stopCh:
				return
			}
		}
		
		// Generate a record
		record := g.generateRecord(recordID)
		
		// Try to send the record (this will block if the buffer is full, providing backpressure)
		select {
		case g.recordCh <- record:
			recordID++
		case <-g.stopCh:
			return
		}
	}
}

// generateRecord creates a single record based on the schema.
func (g *StreamGenerator) generateRecord(recordID int64) types.Record {
	record := make(types.Record)
	
	// Generate the key field first if specified
	if g.schema.Key != "" {
		record[g.schema.Key] = g.generateValue(g.schema.Key, g.schema.Fields[g.schema.Key], recordID)
	}
	
	// Generate values for all other fields
	for fieldName, field := range g.schema.Fields {
		if fieldName == g.schema.Key {
			continue // Already handled above
		}
		record[fieldName] = g.generateValue(fieldName, field, recordID)
	}
	
	return record
}

// generateValue generates a realistic value for a specific field.
func (g *StreamGenerator) generateValue(fieldName string, field *schema.Field, recordID int64) interface{} {
	if field == nil {
		return nil
	}
	
	// Check if there's a custom pattern for this field
	if pattern, exists := g.config.DataPatterns[fieldName]; exists {
		return g.generateFromPattern(pattern, field.Type, recordID)
	}
	
	// Generate based on field type
	switch field.Type {
	case "numeric":
		return g.generateNumeric(fieldName, recordID)
	case "string":
		return g.generateString(fieldName, recordID)
	case "datetime", "date", "timestamp":
		return g.generateDateTime(fieldName, recordID)
	case "boolean":
		return g.rng.Float32() < 0.5
	case "object":
		return g.generateObject(fieldName, recordID)
	case "array":
		return g.generateArray(fieldName, recordID)
	default:
		return g.generateString(fieldName, recordID)
	}
}

// generateFromPattern generates a value based on a custom data pattern.
func (g *StreamGenerator) generateFromPattern(pattern config.DataPattern, fieldType string, recordID int64) interface{} {
	switch pattern.Type {
	case "list":
		if len(pattern.Values) == 0 {
			return nil
		}
		return pattern.Values[g.rng.Intn(len(pattern.Values))]
		
	case "range":
		return g.generateRangeValue(pattern.Min, pattern.Max, fieldType)
		
	case "format":
		return g.generateFormattedValue(pattern.Format, recordID)
		
	default:
		// Fall back to type-based generation
		return g.generateString("", recordID)
	}
}

// generateNumeric generates realistic numeric values.
func (g *StreamGenerator) generateNumeric(fieldName string, recordID int64) interface{} {
	// Generate different patterns based on field name hints
	switch {
	case containsAny(fieldName, []string{"id", "ID", "_id"}):
		return recordID
	case containsAny(fieldName, []string{"age"}):
		return g.rng.Intn(80) + 18 // Ages between 18-98
	case containsAny(fieldName, []string{"price", "cost", "amount"}):
		return float64(g.rng.Intn(100000)) / 100.0 // Prices with cents
	case containsAny(fieldName, []string{"count", "quantity"}):
		return g.rng.Intn(1000) + 1
	default:
		// Random float between 0 and 1000
		return float64(g.rng.Intn(100000)) / 100.0
	}
}

// generateString generates realistic string values.
func (g *StreamGenerator) generateString(fieldName string, recordID int64) interface{} {
	// Generate different patterns based on field name hints
	switch {
	case containsAny(fieldName, []string{"email", "mail"}):
		domains := []string{"example.com", "test.com", "email.com", "domain.org"}
		return fmt.Sprintf("user%d@%s", recordID, domains[g.rng.Intn(len(domains))])
		
	case containsAny(fieldName, []string{"name", "username", "user"}):
		names := []string{"Alice", "Bob", "Charlie", "David", "Eve", "Frank", "Grace", "Henry", "Ivy", "Jack"}
		return names[g.rng.Intn(len(names))]
		
	case containsAny(fieldName, []string{"city", "location"}):
		cities := []string{"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia", "San Antonio", "San Diego", "Dallas", "San Jose"}
		return cities[g.rng.Intn(len(cities))]
		
	case containsAny(fieldName, []string{"plan", "type", "category"}):
		types := []string{"basic", "premium", "enterprise", "free", "standard", "deluxe"}
		return types[g.rng.Intn(len(types))]
		
	case containsAny(fieldName, []string{"status", "state"}):
		statuses := []string{"active", "inactive", "pending", "completed", "failed", "processing"}
		return statuses[g.rng.Intn(len(statuses))]
		
	default:
		// Generate a random alphanumeric string
		return g.generateRandomString(8 + g.rng.Intn(16))
	}
}

// generateDateTime generates realistic datetime values.
func (g *StreamGenerator) generateDateTime(fieldName string, recordID int64) interface{} {
	now := time.Now()
	
	switch {
	case containsAny(fieldName, []string{"created", "created_at", "created_date"}):
		// Random date within the last year
		days := g.rng.Intn(365)
		return now.AddDate(0, 0, -days).Format(time.RFC3339)
		
	case containsAny(fieldName, []string{"updated", "modified", "last_"}):
		// Random date within the last month
		days := g.rng.Intn(30)
		return now.AddDate(0, 0, -days).Format(time.RFC3339)
		
	case containsAny(fieldName, []string{"birth", "dob"}):
		// Random birthdate (18-80 years ago)
		years := g.rng.Intn(62) + 18
		return now.AddDate(-years, -g.rng.Intn(12), -g.rng.Intn(365)).Format("2006-01-02")
		
	default:
		// Random date within the last 6 months
		days := g.rng.Intn(180)
		return now.AddDate(0, 0, -days).Format(time.RFC3339)
	}
}

// generateObject generates a simple map object.
func (g *StreamGenerator) generateObject(fieldName string, recordID int64) interface{} {
	obj := make(map[string]interface{})
	
	// Generate 2-5 fields in the object
	numFields := 2 + g.rng.Intn(4)
	for i := 0; i < numFields; i++ {
		key := fmt.Sprintf("field%d", i+1)
		obj[key] = g.generateString(key, recordID)
	}
	
	return obj
}

// generateArray generates a simple array.
func (g *StreamGenerator) generateArray(fieldName string, recordID int64) interface{} {
	// Generate 1-5 items in the array
	numItems := 1 + g.rng.Intn(5)
	arr := make([]interface{}, numItems)
	
	for i := 0; i < numItems; i++ {
		arr[i] = g.generateString(fieldName, recordID+int64(i))
	}
	
	return arr
}

// Helper functions

func (g *StreamGenerator) generateRangeValue(min, max interface{}, fieldType string) interface{} {
	switch fieldType {
	case "numeric":
		minVal, _ := strconv.ParseFloat(fmt.Sprintf("%v", min), 64)
		maxVal, _ := strconv.ParseFloat(fmt.Sprintf("%v", max), 64)
		if maxVal <= minVal {
			return minVal
		}
		return minVal + g.rng.Float64()*(maxVal-minVal)
	default:
		return fmt.Sprintf("%v", min)
	}
}

func (g *StreamGenerator) generateFormattedValue(format string, recordID int64) interface{} {
	// Simple format substitution
	switch format {
	case "email":
		domains := []string{"example.com", "test.com", "company.com", "email.org", "demo.net"}
		usernames := []string{"user", "test", "admin", "customer", "demo", "sample"}
		username := usernames[g.rng.Intn(len(usernames))]
		domain := domains[g.rng.Intn(len(domains))]
		return fmt.Sprintf("%s%d@%s", username, recordID, domain)
	case "phone":
		return fmt.Sprintf("+1-%03d-%03d-%04d", 
			200+g.rng.Intn(800), g.rng.Intn(900)+100, g.rng.Intn(10000))
	case "uuid":
		return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			g.rng.Uint32(), g.rng.Uint32()&0xffff, g.rng.Uint32()&0xffff,
			g.rng.Uint32()&0xffff, g.rng.Uint64()&0xffffffffffff)
	case "ip":
		return fmt.Sprintf("%d.%d.%d.%d", 
			10+g.rng.Intn(245), g.rng.Intn(256), g.rng.Intn(256), 1+g.rng.Intn(254))
	case "mac":
		return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
			g.rng.Intn(256), g.rng.Intn(256), g.rng.Intn(256),
			g.rng.Intn(256), g.rng.Intn(256), g.rng.Intn(256))
	case "api_key":
		return fmt.Sprintf("ak_%s", g.generateRandomString(32))
	default:
		// Handle dynamic format patterns
		if strings.Contains(format, "{id}") {
			return strings.ReplaceAll(format, "{id}", strconv.FormatInt(recordID, 10))
		}
		if strings.Contains(format, "{random}") {
			return strings.ReplaceAll(format, "{random}", g.generateRandomString(8))
		}
		return format
	}
}

func (g *StreamGenerator) generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[g.rng.Intn(len(charset))]
	}
	return string(result)
}

func containsAny(str string, substrings []string) bool {
	for _, substring := range substrings {
		if len(str) >= len(substring) {
			for i := 0; i <= len(str)-len(substring); i++ {
				if str[i:i+len(substring)] == substring {
					return true
				}
			}
		}
	}
	return false
}