package comparator

import (
	"data-comparator/internal/pkg/config"
	"data-comparator/internal/pkg/datareader"
	"fmt"
	"io"
	"time"
)

// ComparisonResult holds the results of comparing two data streams.
type ComparisonResult struct {
	Timestamp            time.Time            `yaml:"timestamp"`
	RecordsProcessed     int                  `yaml:"records_processed"`
	Source1Records       int                  `yaml:"source1_records"`
	Source2Records       int                  `yaml:"source2_records"`
	MatchingKeys         int                  `yaml:"matching_keys"`
	IdenticalRows        int                  `yaml:"identical_rows"`
	KeysOnlyInSource1    []string            `yaml:"keys_only_in_source1,omitempty"`
	KeysOnlyInSource2    []string            `yaml:"keys_only_in_source2,omitempty"`
	ValueDiffs           map[string][]FieldDiff `yaml:"value_diffs_by_key,omitempty"`
	IsPeriodicReport     bool                 `yaml:"is_periodic_report"`
}

// FieldDiff represents a difference in a field between two records.
type FieldDiff struct {
	Field         string      `yaml:"field"`
	Source1Value  interface{} `yaml:"source1_value"`
	Source2Value  interface{} `yaml:"source2_value"`
}

// StreamComparator compares two data streams with periodic reporting.
type StreamComparator struct {
	source1         datareader.DataReader
	source2         datareader.DataReader
	periodicConfig  config.PeriodicConfig
	onPeriodicDiff  func(result ComparisonResult) error
	keyField        string
}

// NewStreamComparator creates a new stream comparator.
func NewStreamComparator(
	source1, source2 datareader.DataReader,
	periodicConfig config.PeriodicConfig,
	keyField string,
	onPeriodicDiff func(result ComparisonResult) error,
) *StreamComparator {
	return &StreamComparator{
		source1:        source1,
		source2:        source2,
		periodicConfig: periodicConfig,
		onPeriodicDiff: onPeriodicDiff,
		keyField:       keyField,
	}
}

// Compare performs the stream comparison with periodic reporting.
func (sc *StreamComparator) Compare() (*ComparisonResult, error) {
	startTime := time.Now()
	lastReportTime := startTime
	recordsProcessed := 0
	lastReportRecords := 0

	// Maps to store records by key
	source1Records := make(map[string]datareader.Record)
	source2Records := make(map[string]datareader.Record)

	source1Count, source2Count := 0, 0
	
	// Read all records from source1
	for {
		record, err := sc.source1.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error reading from source1: %w", err)
		}
		
		source1Count++
		recordsProcessed++
		
		key := sc.getRecordKey(record)
		if key != "" {
			source1Records[key] = record
		}

		// Check for periodic reporting
		if sc.shouldReportPeriodic(startTime, lastReportTime, recordsProcessed, lastReportRecords) {
			result := sc.generatePeriodicResult(source1Records, source2Records, recordsProcessed, source1Count, source2Count)
			if sc.onPeriodicDiff != nil {
				if err := sc.onPeriodicDiff(result); err != nil {
					return nil, fmt.Errorf("error in periodic diff callback: %w", err)
				}
			}
			lastReportTime = time.Now()
			lastReportRecords = recordsProcessed
		}
	}

	// Read all records from source2
	for {
		record, err := sc.source2.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error reading from source2: %w", err)
		}
		
		source2Count++
		recordsProcessed++
		
		key := sc.getRecordKey(record)
		if key != "" {
			source2Records[key] = record
		}

		// Check for periodic reporting
		if sc.shouldReportPeriodic(startTime, lastReportTime, recordsProcessed, lastReportRecords) {
			result := sc.generatePeriodicResult(source1Records, source2Records, recordsProcessed, source1Count, source2Count)
			if sc.onPeriodicDiff != nil {
				if err := sc.onPeriodicDiff(result); err != nil {
					return nil, fmt.Errorf("error in periodic diff callback: %w", err)
				}
			}
			lastReportTime = time.Now()
			lastReportRecords = recordsProcessed
		}
	}

	// Generate final result
	return sc.generateFinalResult(source1Records, source2Records, recordsProcessed, source1Count, source2Count), nil
}

// getRecordKey extracts the key field value from a record.
func (sc *StreamComparator) getRecordKey(record datareader.Record) string {
	if sc.keyField == "" {
		return ""
	}
	
	value, exists := record[sc.keyField]
	if !exists {
		return ""
	}
	
	return fmt.Sprintf("%v", value)
}

// shouldReportPeriodic determines if a periodic report should be generated.
func (sc *StreamComparator) shouldReportPeriodic(startTime, lastReportTime time.Time, recordsProcessed, lastReportRecords int) bool {
	if !sc.periodicConfig.Enabled {
		return false
	}

	// Check time interval
	if sc.periodicConfig.TimeInterval > 0 {
		if time.Since(lastReportTime).Seconds() >= float64(sc.periodicConfig.TimeInterval) {
			return true
		}
	}

	// Check record interval
	if sc.periodicConfig.RecordInterval > 0 {
		if recordsProcessed-lastReportRecords >= sc.periodicConfig.RecordInterval {
			return true
		}
	}

	return false
}

// generatePeriodicResult creates a periodic comparison result.
func (sc *StreamComparator) generatePeriodicResult(
	source1Records, source2Records map[string]datareader.Record,
	recordsProcessed, source1Count, source2Count int,
) ComparisonResult {
	return sc.generateResult(source1Records, source2Records, recordsProcessed, source1Count, source2Count, true)
}

// generateFinalResult creates the final comparison result.
func (sc *StreamComparator) generateFinalResult(
	source1Records, source2Records map[string]datareader.Record,
	recordsProcessed, source1Count, source2Count int,
) *ComparisonResult {
	result := sc.generateResult(source1Records, source2Records, recordsProcessed, source1Count, source2Count, false)
	return &result
}

// generateResult creates a comparison result.
func (sc *StreamComparator) generateResult(
	source1Records, source2Records map[string]datareader.Record,
	recordsProcessed, source1Count, source2Count int,
	isPeriodicReport bool,
) ComparisonResult {
	result := ComparisonResult{
		Timestamp:         time.Now(),
		RecordsProcessed:  recordsProcessed,
		Source1Records:    source1Count,
		Source2Records:    source2Count,
		IsPeriodicReport:  isPeriodicReport,
		ValueDiffs:        make(map[string][]FieldDiff),
	}

	// Find keys only in source1
	for key := range source1Records {
		if _, exists := source2Records[key]; !exists {
			result.KeysOnlyInSource1 = append(result.KeysOnlyInSource1, key)
		}
	}

	// Find keys only in source2
	for key := range source2Records {
		if _, exists := source1Records[key]; !exists {
			result.KeysOnlyInSource2 = append(result.KeysOnlyInSource2, key)
		}
	}

	// Compare records with matching keys
	for key, record1 := range source1Records {
		if record2, exists := source2Records[key]; exists {
			result.MatchingKeys++
			
			// Compare field values
			diffs := sc.compareRecords(record1, record2)
			if len(diffs) == 0 {
				result.IdenticalRows++
			} else {
				result.ValueDiffs[key] = diffs
			}
		}
	}

	return result
}

// compareRecords compares two records and returns field differences.
func (sc *StreamComparator) compareRecords(record1, record2 datareader.Record) []FieldDiff {
	var diffs []FieldDiff

	// Get all unique field names
	allFields := make(map[string]bool)
	for field := range record1 {
		allFields[field] = true
	}
	for field := range record2 {
		allFields[field] = true
	}

	// Compare each field
	for field := range allFields {
		value1, exists1 := record1[field]
		value2, exists2 := record2[field]

		if !exists1 && !exists2 {
			continue
		}

		if !exists1 || !exists2 || !sc.valuesEqual(value1, value2) {
			diffs = append(diffs, FieldDiff{
				Field:        field,
				Source1Value: value1,
				Source2Value: value2,
			})
		}
	}

	return diffs
}

// valuesEqual compares two values for equality.
func (sc *StreamComparator) valuesEqual(v1, v2 interface{}) bool {
	if v1 == nil && v2 == nil {
		return true
	}
	if v1 == nil || v2 == nil {
		return false
	}
	
	// Simple string comparison for now
	str1 := fmt.Sprintf("%v", v1)
	str2 := fmt.Sprintf("%v", v2)
	return str1 == str2
}