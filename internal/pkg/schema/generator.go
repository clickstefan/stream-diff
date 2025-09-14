package schema

import (
	"data-comparator/internal/pkg/config"
	"data-comparator/internal/pkg/datareader"
	"fmt"
	"io"
	"strconv"
	"time"
)

// DefaultSampleSize is the number of records to sample if not specified in the config.
const DefaultSampleSize = 1000

// Generate creates a schema by sampling records from a data reader.
func Generate(reader datareader.DataReader, samplerConfig *config.Sampler) (*Schema, error) {
	sampleSize := DefaultSampleSize
	if samplerConfig != nil && samplerConfig.SampleSize > 0 {
		sampleSize = samplerConfig.SampleSize
	}

	records, err := sampleRecords(reader, sampleSize)
	if err != nil {
		return nil, fmt.Errorf("failed to sample records: %w", err)
	}
	if len(records) == 0 {
		return &Schema{Fields: make(map[string]*Field)}, nil
	}

	fieldValues := make(map[string][]interface{})
	for _, record := range records {
		CollectFieldValues(record, fieldValues)
	}

	fields := analyzeFields(fieldValues)
	schema := &Schema{
		Fields: fields,
	}

	// TODO: Implement key identification
	return schema, nil
}

func analyzeFields(fieldValues map[string][]interface{}) map[string]*Field {
	fields := make(map[string]*Field)
	for name, values := range fieldValues {
		fields[name] = &Field{
			Type:  inferType(values),
			Stats: []string{}, // TODO: Calculate stats based on type
		}
	}
	return fields
}

func inferType(values []interface{}) string {
	if len(values) == 0 {
		return "unknown"
	}
	isNumeric, isDateTime, isObject, isArray := true, true, true, true
	dateTimeLayouts := []string{
		time.RFC3339, time.RFC3339Nano, "2006-01-02 15:04:05", "2006-01-02", "01/02/2006",
	}
	nonNilCount := 0
	for _, val := range values {
		if val == nil {
			continue
		}
		nonNilCount++
		if _, ok := val.(map[string]interface{}); !ok {
			isObject = false
		}
		if _, ok := val.([]interface{}); !ok {
			isArray = false
		}
		sVal := fmt.Sprintf("%v", val)
		if _, err := strconv.ParseFloat(sVal, 64); err != nil {
			isNumeric = false
		}
		canBeDateTime := false
		for _, layout := range dateTimeLayouts {
			if _, err := time.Parse(layout, sVal); err == nil {
				canBeDateTime = true
				break
			}
		}
		if !canBeDateTime {
			isDateTime = false
		}
	}
	if nonNilCount == 0 {
		return "unknown"
	}
	if isObject {
		return "object"
	}
	if isArray {
		return "array"
	}
	if isNumeric {
		return "numeric"
	}
	if isDateTime {
		return "datetime"
	}
	return "string"
}

func sampleRecords(reader datareader.DataReader, sampleSize int) ([]datareader.Record, error) {
	var records []datareader.Record
	for i := 0; i < sampleSize; i++ {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

type workItem struct {
	data   interface{}
	prefix string
}

func CollectFieldValues(data interface{}, fieldValues map[string][]interface{}) {
	queue := []workItem{{data: data, prefix: ""}}

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		if item.data == nil {
			continue
		}

		if m, ok := item.data.(map[string]interface{}); ok {
			if item.prefix != "" {
				fieldValues[item.prefix] = append(fieldValues[item.prefix], m)
			}
			for key, value := range m {
				newKey := key
				if item.prefix != "" {
					newKey = item.prefix + "." + key
				}
				queue = append(queue, workItem{data: value, prefix: newKey})
			}
		} else if a, ok := item.data.([]interface{}); ok {
			if item.prefix != "" {
				fieldValues[item.prefix] = append(fieldValues[item.prefix], a)
			}
			arrayKey := item.prefix + "[]"
			for _, v := range a {
				queue = append(queue, workItem{data: v, prefix: arrayKey})
			}
		} else {
			// Restore the prefix check, as removing it was a failed diagnostic
			if item.prefix != "" {
				fieldValues[item.prefix] = append(fieldValues[item.prefix], item.data)
			}
		}
	}
}
