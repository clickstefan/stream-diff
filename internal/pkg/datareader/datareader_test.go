package datareader

import (
	"data-comparator/internal/pkg/config"
	"io"
	"reflect"
	"testing"
)

func TestCSVReader_Simple(t *testing.T) {
	cfg := config.Source{
		Type: "csv",
		Path: "../../../testdata/testcase1_simple_csv/source1.csv",
	}
	reader, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer reader.Close()

	// Read first record
	rec, err := reader.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	expected := Record{
		"user_id":    "1",
		"email":      "alice@email.com",
		"age":        "30",
		"city":       "New York",
		"plan_type":  "premium",
		"last_login": "2025-09-10T12:00:00Z",
	}

	if !reflect.DeepEqual(rec, expected) {
		t.Errorf("Read() got = %v, want %v", rec, expected)
	}
}

func TestJSONReader(t *testing.T) {
	cfg := config.Source{
		Type: "json",
		Path: "../../../testdata/testcase2_nested_json/source1.jsonl",
	}
	reader, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer reader.Close()

	rec, err := reader.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	// Check a nested value
	if customer, ok := rec["customer"].(map[string]interface{}); ok {
		if region, ok := customer["region"].(string); ok {
			if region != "us-east-1" {
				t.Errorf("Nested field read incorrectly, got %s, want %s", region, "us-east-1")
			}
		} else {
			t.Error("Nested field 'region' is not a string")
		}
	} else {
		t.Error("Field 'customer' is not a map")
	}
}

func TestCSVReader_WithEmbeddedJSON(t *testing.T) {
	cfg := config.Source{
		Type: "csv",
		Path: "../../../testdata/testcase3_csv_with_json/source1.csv",
		ParserConfig: &config.ParserConfig{
			JSONInString: true,
		},
	}
	reader, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer reader.Close()

	rec, err := reader.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	// Check that the 'details' field was parsed into a map
	details, ok := rec["details"].(map[string]interface{})
	if !ok {
		t.Fatalf("'details' field was not parsed as JSON object, type is %T", rec["details"])
	}

	// Check that the 'payload' field within 'details' was also parsed (recursively)
	payload, ok := details["payload"].(map[string]interface{})
	if !ok {
		t.Fatalf("'payload' field was not parsed as JSON object, type is %T", details["payload"])
	}

	// Check a value in the innermost JSON object
	if source, ok := payload["source"].(string); ok {
		if source != "web" {
			t.Errorf("Innermost field 'source' read incorrectly, got %s, want %s", source, "web")
		}
	} else {
		t.Error("Innermost field 'source' is not a string")
	}
}

func TestReader_EOF(t *testing.T) {
	cfg := config.Source{
		Type: "csv",
		Path: "../../../testdata/testcase1_simple_csv/source1.csv",
	}
	reader, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer reader.Close()

	// Read all 5 records
	for i := 0; i < 5; i++ {
		_, err := reader.Read()
		if err != nil {
			t.Fatalf("Read() error on record %d: %v", i+1, err)
		}
	}

	// The next read should return io.EOF
	_, err = reader.Read()
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}
