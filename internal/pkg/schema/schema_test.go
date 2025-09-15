package schema

import (
	"data-comparator/internal/pkg/config"
	"data-comparator/internal/pkg/datareader"
	"reflect"
	"testing"
)

func TestGenerate_SimpleCSV(t *testing.T) {
	cfg, err := config.Load("../../../testdata/testcase1_simple_csv/config1.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	cfg.Source.Path = "../../../" + cfg.Source.Path

	reader, err := datareader.New(cfg.Source)
	if err != nil {
		t.Fatalf("Failed to create data reader: %v", err)
	}
	defer reader.Close()

	schema, err := Generate(reader, cfg.Source.Sampler)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if schema == nil {
		t.Fatal("Schema is nil")
	}

	// For testcase1, there are 6 fields
	if len(schema.Fields) != 6 {
		t.Fatalf("Expected 6 fields, but got %d. Fields: %v", len(schema.Fields), reflect.ValueOf(schema.Fields).MapKeys())
	}

	expectedTypes := map[string]string{
		"user_id":    "numeric",
		"email":      "string",
		"age":        "numeric",
		"city":       "string",
		"plan_type":  "string",
		"last_login": "datetime",
	}

	for fieldName, expectedType := range expectedTypes {
		field, ok := schema.Fields[fieldName]
		if !ok {
			t.Errorf("Expected field '%s' not found in schema", fieldName)
			continue
		}
		if field.Type != expectedType {
			t.Errorf("Field '%s' has wrong type: got %s, want %s", fieldName, field.Type, expectedType)
		}
	}
}

func TestCollectFieldValues(t *testing.T) {
	record := map[string]interface{}{
		"id":   float64(1),
		"user": map[string]interface{}{"name": "Jules"},
		"tags": []interface{}{"go", "test"},
	}
	fieldValues := make(map[string][]interface{})
	CollectFieldValues(record, fieldValues)

	expectedKeys := []string{"id", "user", "user.name", "tags", "tags[]"}
	for _, k := range expectedKeys {
		if _, ok := fieldValues[k]; !ok {
			t.Errorf("Expected key '%s' not found", k)
		}
	}

	if len(fieldValues) != len(expectedKeys) {
		t.Errorf("Expected %d fields, got %d. Keys: %v", len(expectedKeys), len(fieldValues), reflect.ValueOf(fieldValues).MapKeys())
	}
}
