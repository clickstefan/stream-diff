package schema

import (
	"data-comparator/internal/pkg/types"
	"reflect"
	"testing"
)

func TestGenerate_SimpleCSV(t *testing.T) {
	// Create test data similar to testcase1
	testRecords := []types.Record{
		{"user_id": "1", "email": "alice@email.com", "age": "30", "city": "New York", "plan_type": "premium", "last_login": "2025-09-10T12:00:00Z"},
		{"user_id": "2", "email": "bob@email.com", "age": "25", "city": "Los Angeles", "plan_type": "basic", "last_login": "2025-09-11 10:00:00"},
		{"user_id": "3", "email": "charlie@email.com", "age": "35", "city": "Chicago", "plan_type": "premium", "last_login": "09/12/2025"},
		{"user_id": "4", "email": "david@email.com", "age": "40", "city": "New York", "plan_type": "basic", "last_login": "2025-09-12"},
		{"user_id": "5", "email": "eve@email.com", "age": "28", "city": "Chicago", "plan_type": "basic", "last_login": "2025-09-13T05:30:00+01:00"},
	}
	
	reader := newTestReader(testRecords)
	defer reader.Close()

	schema, err := Generate(reader, nil)
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
