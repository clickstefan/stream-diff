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

func TestProtobufReader_JSONFormat(t *testing.T) {
	cfg := config.Source{
		Type: "protobuf",
		Path: "../../../testdata/testcase4_protobuf/source1.jsonpb",
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

	// Check basic field
	if userID, ok := rec["user_id"].(string); ok {
		if userID != "user-001" {
			t.Errorf("user_id read incorrectly, got %s, want %s", userID, "user-001")
		}
	} else {
		t.Error("Field 'user_id' is not a string")
	}

	// Check nested field
	if profile, ok := rec["profile"].(map[string]interface{}); ok {
		if email, ok := profile["email"].(string); ok {
			if email != "alice@example.com" {
				t.Errorf("Nested field email read incorrectly, got %s, want %s", email, "alice@example.com")
			}
		} else {
			t.Error("Nested field 'email' is not a string")
		}
	} else {
		t.Error("Field 'profile' is not a map")
	}

	// Check array field
	if activity, ok := rec["activity"].(map[string]interface{}); ok {
		if devices, ok := activity["devices"].([]interface{}); ok {
			if len(devices) != 2 {
				t.Errorf("devices array length incorrect, got %d, want %d", len(devices), 2)
			}
			if devices[0].(string) != "mobile" {
				t.Errorf("First device incorrect, got %s, want %s", devices[0], "mobile")
			}
		} else {
			t.Error("Field 'devices' is not an array")
		}
	} else {
		t.Error("Field 'activity' is not a map")
	}
}

func TestProtobufReader_MultipleTypes(t *testing.T) {
	// Test both "protobuf" and "proto" as valid type names
	for _, sourceType := range []string{"protobuf", "proto"} {
		t.Run(sourceType, func(t *testing.T) {
			cfg := config.Source{
				Type: sourceType,
				Path: "../../../testdata/testcase4_protobuf/source1.jsonpb",
			}
			reader, err := New(cfg)
			if err != nil {
				t.Fatalf("New() error for type %s = %v", sourceType, err)
			}
			defer reader.Close()

			// Just verify we can read one record without error
			_, err = reader.Read()
			if err != nil {
				t.Fatalf("Read() error for type %s = %v", sourceType, err)
			}
		})
	}
}

func TestProtobufReader_EOF(t *testing.T) {
	cfg := config.Source{
		Type: "protobuf",
		Path: "../../../testdata/testcase4_protobuf/source1.jsonpb",
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
