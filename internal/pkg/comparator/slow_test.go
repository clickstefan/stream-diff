package comparator

import (
	"data-comparator/internal/pkg/config"
	"data-comparator/internal/pkg/datareader"
	"fmt"
	"io"
	"time"
)

// SlowMockDataReader simulates slow data reading for testing time-based triggers
type SlowMockDataReader struct {
	records   []datareader.Record
	index     int
	delay     time.Duration
}

func NewSlowMockDataReader(records []datareader.Record, delay time.Duration) *SlowMockDataReader {
	return &SlowMockDataReader{
		records: records,
		index:   0,
		delay:   delay,
	}
}

func (m *SlowMockDataReader) Read() (datareader.Record, error) {
	if m.index >= len(m.records) {
		return nil, io.EOF
	}
	
	// Add delay to simulate slow reading
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	
	record := m.records[m.index]
	m.index++
	return record, nil
}

func (m *SlowMockDataReader) Close() error {
	return nil
}

// CreateSlowReaderTest creates a test that demonstrates time-based periodic reporting
func CreateSlowReaderTest() {
	fmt.Println("Testing time-based periodic reporting...")
	
	// Create test data with enough records to span time intervals
	source1Records := []datareader.Record{
		{"id": "1", "name": "Alice", "age": "30"},
		{"id": "2", "name": "Bob", "age": "25"},
		{"id": "3", "name": "Charlie", "age": "35"},
		{"id": "4", "name": "David", "age": "40"},
		{"id": "5", "name": "Eve", "age": "28"},
	}

	source2Records := []datareader.Record{
		{"id": "1", "name": "Alice", "age": "31"},
		{"id": "2", "name": "Bob", "age": "25"},
		{"id": "3", "name": "Charlie", "age": "36"},
		{"id": "6", "name": "Frank", "age": "45"},
		{"id": "7", "name": "Grace", "age": "32"},
	}

	// Create slow readers with 2-second delay per record
	reader1 := NewSlowMockDataReader(source1Records, 2*time.Second)
	reader2 := NewSlowMockDataReader(source2Records, 2*time.Second)

	// Configure for time-based periodic reporting every 3 seconds
	periodicConfig := config.PeriodicConfig{
		Enabled:      true,
		TimeInterval: 3, // 3 seconds
		RecordInterval: 0, // Disable record-based trigger
	}
	
	// Track periodic reports
	periodicCallback := func(result ComparisonResult) error {
		fmt.Printf("[PERIODIC TIME-BASED] %s - Records: %d, Matching: %d\n",
			result.Timestamp.Format("15:04:05"),
			result.RecordsProcessed,
			result.MatchingKeys)
		return nil
	}
	
	sc := NewStreamComparator(reader1, reader2, periodicConfig, "id", periodicCallback)

	// Perform comparison
	fmt.Println("Starting slow comparison to demonstrate time-based triggers...")
	start := time.Now()
	
	result, err := sc.Compare()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	duration := time.Since(start)
	fmt.Printf("Completed in %v\n", duration)
	fmt.Printf("Final: Records: %d, Matching: %d, Identical: %d\n",
		result.RecordsProcessed, result.MatchingKeys, result.IdenticalRows)
}