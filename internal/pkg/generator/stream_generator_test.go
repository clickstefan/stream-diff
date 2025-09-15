package generator

import (
	"data-comparator/internal/pkg/config"
	"data-comparator/internal/pkg/schema"
	"io"
	"testing"
	"time"
)

func TestStreamGenerator_BasicGeneration(t *testing.T) {
	// Create a simple schema
	testSchema := &schema.Schema{
		Key: "user_id",
		Fields: map[string]*schema.Field{
			"user_id": {Type: "numeric"},
			"email":   {Type: "string"},
			"age":     {Type: "numeric"},
			"active":  {Type: "boolean"},
		},
	}
	
	// Create generator with basic config
	generatorConfig := config.StreamGeneratorConfig{
		MaxRecords: 5,
		BufferSize: 10,
		Seed:       12345, // Fixed seed for reproducible tests
	}
	
	generator, err := New(testSchema, generatorConfig)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	defer generator.Close()
	
	// Read all records
	var records []map[string]interface{}
	for i := 0; i < 5; i++ {
		record, err := generator.Read()
		if err != nil {
			t.Fatalf("Failed to read record %d: %v", i+1, err)
		}
		records = append(records, record)
	}
	
	// Should get EOF on next read
	_, err = generator.Read()
	if err != io.EOF {
		t.Errorf("Expected EOF after max records, got: %v", err)
	}
	
	// Validate records
	if len(records) != 5 {
		t.Fatalf("Expected 5 records, got %d", len(records))
	}
	
	for i, record := range records {
		// Check that all fields are present
		if len(record) != 4 {
			t.Errorf("Record %d should have 4 fields, got %d: %v", i+1, len(record), record)
		}
		
		// Check user_id is sequential
		userID, ok := record["user_id"]
		if !ok {
			t.Errorf("Record %d missing user_id", i+1)
			continue
		}
		if userID != int64(i+1) {
			t.Errorf("Record %d user_id should be %d, got %v", i+1, i+1, userID)
		}
		
		// Check email format
		email, ok := record["email"].(string)
		if !ok || email == "" {
			t.Errorf("Record %d should have non-empty email string, got %v", i+1, record["email"])
		}
		
		// Check boolean field
		_, ok = record["active"].(bool)
		if !ok {
			t.Errorf("Record %d active field should be boolean, got %v", i+1, record["active"])
		}
	}
}

func TestStreamGenerator_RateLimiting(t *testing.T) {
	testSchema := &schema.Schema{
		Fields: map[string]*schema.Field{
			"id": {Type: "numeric"},
		},
	}
	
	// Create generator with rate limiting (10 records per second)
	generatorConfig := config.StreamGeneratorConfig{
		MaxRecords:       3,
		RecordsPerSecond: 10.0,
		BufferSize:       10,
		Seed:             12345,
	}
	
	generator, err := New(testSchema, generatorConfig)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	defer generator.Close()
	
	start := time.Now()
	
	// Read all records
	for i := 0; i < 3; i++ {
		_, err := generator.Read()
		if err != nil {
			t.Fatalf("Failed to read record %d: %v", i+1, err)
		}
	}
	
	elapsed := time.Since(start)
	
	// Should take at least 200ms for 3 records at 10/second (200ms between records)
	expectedMinDuration := 200 * time.Millisecond
	if elapsed < expectedMinDuration {
		t.Errorf("Expected at least %v for rate limiting, but took %v", expectedMinDuration, elapsed)
	}
}

func TestStreamGenerator_DataPatterns(t *testing.T) {
	testSchema := &schema.Schema{
		Fields: map[string]*schema.Field{
			"status": {Type: "string"},
			"score":  {Type: "numeric"},
		},
	}
	
	// Create generator with custom data patterns
	generatorConfig := config.StreamGeneratorConfig{
		MaxRecords: 10,
		BufferSize: 10,
		Seed:       12345,
		DataPatterns: map[string]config.DataPattern{
			"status": {
				Type:   "list",
				Values: []interface{}{"active", "inactive", "pending"},
			},
			"score": {
				Type: "range",
				Min:  0.0,
				Max:  100.0,
			},
		},
	}
	
	generator, err := New(testSchema, generatorConfig)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	defer generator.Close()
	
	// Read some records and verify patterns
	validStatuses := map[string]bool{"active": true, "inactive": true, "pending": true}
	
	for i := 0; i < 5; i++ {
		record, err := generator.Read()
		if err != nil {
			t.Fatalf("Failed to read record %d: %v", i+1, err)
		}
		
		// Check status is from the list
		status, ok := record["status"].(string)
		if !ok {
			t.Errorf("Record %d status should be string, got %v", i+1, record["status"])
			continue
		}
		if !validStatuses[status] {
			t.Errorf("Record %d status %s not in allowed list", i+1, status)
		}
		
		// Check score is in range
		score, ok := record["score"].(float64)
		if !ok {
			t.Errorf("Record %d score should be float64, got %v", i+1, record["score"])
			continue
		}
		if score < 0.0 || score > 100.0 {
			t.Errorf("Record %d score %f should be in range 0-100", i+1, score)
		}
	}
}

func TestStreamGenerator_Backpressure(t *testing.T) {
	testSchema := &schema.Schema{
		Fields: map[string]*schema.Field{
			"id": {Type: "numeric"},
		},
	}
	
	// Create generator with small buffer
	generatorConfig := config.StreamGeneratorConfig{
		MaxRecords: 100,
		BufferSize: 2, // Very small buffer to test backpressure
		Seed:       12345,
	}
	
	generator, err := New(testSchema, generatorConfig)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	defer generator.Close()
	
	// Read records slowly to ensure backpressure works
	for i := 0; i < 5; i++ {
		record, err := generator.Read()
		if err != nil {
			t.Fatalf("Failed to read record %d: %v", i+1, err)
		}
		if record == nil {
			t.Errorf("Record %d should not be nil", i+1)
		}
		
		// Small delay to test that generator doesn't overflow
		time.Sleep(10 * time.Millisecond)
	}
}

func TestStreamGenerator_Close(t *testing.T) {
	testSchema := &schema.Schema{
		Fields: map[string]*schema.Field{
			"id": {Type: "numeric"},
		},
	}
	
	generatorConfig := config.StreamGeneratorConfig{
		MaxRecords: 1000, // More than we'll read
		BufferSize: 10,
		Seed:       12345,
	}
	
	generator, err := New(testSchema, generatorConfig)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	
	// Read a few records
	for i := 0; i < 3; i++ {
		_, err := generator.Read()
		if err != nil {
			t.Fatalf("Failed to read record %d: %v", i+1, err)
		}
	}
	
	// Close the generator
	err = generator.Close()
	if err != nil {
		t.Fatalf("Failed to close generator: %v", err)
	}
	
	// Next read should return EOF
	_, err = generator.Read()
	if err != io.EOF {
		t.Errorf("Expected EOF after close, got: %v", err)
	}
	
	// Close should be idempotent
	err = generator.Close()
	if err != nil {
		t.Errorf("Second close should not fail: %v", err)
	}
}