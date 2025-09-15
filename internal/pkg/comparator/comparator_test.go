package comparator

import (
	"data-comparator/internal/pkg/config"
	"data-comparator/internal/pkg/datareader"
	"io"
	"testing"
	"time"
)

// MockDataReader is a mock implementation of DataReader for testing
type MockDataReader struct {
	records []datareader.Record
	index   int
}

func NewMockDataReader(records []datareader.Record) *MockDataReader {
	return &MockDataReader{
		records: records,
		index:   0,
	}
}

func (m *MockDataReader) Read() (datareader.Record, error) {
	if m.index >= len(m.records) {
		return nil, io.EOF
	}
	record := m.records[m.index]
	m.index++
	return record, nil
}

func (m *MockDataReader) Close() error {
	return nil
}

func TestStreamComparator_Compare(t *testing.T) {
	// Create test data
	source1Records := []datareader.Record{
		{"id": "1", "name": "Alice", "age": "30"},
		{"id": "2", "name": "Bob", "age": "25"},
		{"id": "3", "name": "Charlie", "age": "35"},
	}

	source2Records := []datareader.Record{
		{"id": "1", "name": "Alice", "age": "31"}, // Age diff
		{"id": "2", "name": "Bob", "age": "25"},   // Identical
		{"id": "4", "name": "David", "age": "40"}, // Only in source2
	}

	// Create mock readers
	reader1 := NewMockDataReader(source1Records)
	reader2 := NewMockDataReader(source2Records)

	// Create comparator without periodic reporting
	periodicConfig := config.PeriodicConfig{
		Enabled: false,
	}
	
	sc := NewStreamComparator(reader1, reader2, periodicConfig, "id", nil)

	// Perform comparison
	result, err := sc.Compare()
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}

	// Verify results
	if result.RecordsProcessed != 6 {
		t.Errorf("RecordsProcessed got = %v, want %v", result.RecordsProcessed, 6)
	}
	if result.Source1Records != 3 {
		t.Errorf("Source1Records got = %v, want %v", result.Source1Records, 3)
	}
	if result.Source2Records != 3 {
		t.Errorf("Source2Records got = %v, want %v", result.Source2Records, 3)
	}
	if result.MatchingKeys != 2 {
		t.Errorf("MatchingKeys got = %v, want %v", result.MatchingKeys, 2)
	}
	if result.IdenticalRows != 1 {
		t.Errorf("IdenticalRows got = %v, want %v", result.IdenticalRows, 1)
	}

	// Check keys only in source1
	if len(result.KeysOnlyInSource1) != 1 || result.KeysOnlyInSource1[0] != "3" {
		t.Errorf("KeysOnlyInSource1 got = %v, want %v", result.KeysOnlyInSource1, []string{"3"})
	}

	// Check keys only in source2
	if len(result.KeysOnlyInSource2) != 1 || result.KeysOnlyInSource2[0] != "4" {
		t.Errorf("KeysOnlyInSource2 got = %v, want %v", result.KeysOnlyInSource2, []string{"4"})
	}

	// Check value diffs
	if len(result.ValueDiffs) != 1 {
		t.Errorf("ValueDiffs length got = %v, want %v", len(result.ValueDiffs), 1)
	}
	
	if diffs, exists := result.ValueDiffs["1"]; exists {
		if len(diffs) != 1 {
			t.Errorf("ValueDiffs for key '1' length got = %v, want %v", len(diffs), 1)
		} else if diffs[0].Field != "age" || diffs[0].Source1Value != "30" || diffs[0].Source2Value != "31" {
			t.Errorf("ValueDiffs for key '1' got = %v, want age diff 30->31", diffs[0])
		}
	} else {
		t.Error("Expected value diff for key '1' not found")
	}
}

func TestStreamComparator_PeriodicReporting(t *testing.T) {
	// Create test data
	source1Records := []datareader.Record{
		{"id": "1", "name": "Alice"},
		{"id": "2", "name": "Bob"},
	}

	source2Records := []datareader.Record{
		{"id": "1", "name": "Alice"},
	}

	// Create mock readers
	reader1 := NewMockDataReader(source1Records)
	reader2 := NewMockDataReader(source2Records)

	// Track periodic reports
	var periodicReports []ComparisonResult
	periodicCallback := func(result ComparisonResult) error {
		periodicReports = append(periodicReports, result)
		return nil
	}

	// Create comparator with periodic reporting every 1 record
	periodicConfig := config.PeriodicConfig{
		Enabled:        true,
		RecordInterval: 1,
	}
	
	sc := NewStreamComparator(reader1, reader2, periodicConfig, "id", periodicCallback)

	// Perform comparison
	_, err := sc.Compare()
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}

	// Should have received periodic reports
	if len(periodicReports) == 0 {
		t.Error("Expected periodic reports but got none")
	}

	// All periodic reports should be marked as such
	for i, report := range periodicReports {
		if !report.IsPeriodicReport {
			t.Errorf("Periodic report %d not marked as periodic", i)
		}
	}
}

func TestStreamComparator_TimeBasedPeriodic(t *testing.T) {
	// This test is more challenging to write without actual time delays
	// For now, we'll just test the configuration
	periodicConfig := config.PeriodicConfig{
		Enabled:      true,
		TimeInterval: 1, // 1 second
	}

	reader1 := NewMockDataReader([]datareader.Record{})
	reader2 := NewMockDataReader([]datareader.Record{})

	sc := NewStreamComparator(reader1, reader2, periodicConfig, "id", nil)

	// Test the shouldReportPeriodic method
	startTime := time.Now()
	lastReportTime := startTime.Add(-2 * time.Second) // 2 seconds ago
	
	should := sc.shouldReportPeriodic(startTime, lastReportTime, 10, 5)
	if !should {
		t.Error("Should report periodic when time interval exceeded")
	}
}