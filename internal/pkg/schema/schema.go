package schema

// Schema represents the learned or defined structure of a data source.
type Schema struct {
	Key        string           `yaml:"key"`
	MaxKeySize int              `yaml:"max_key_size,omitempty"`
	Fields     map[string]*Field `yaml:"fields"`
}

// Field represents the schema for a single field within the data source.
type Field struct {
	Type     string      `yaml:"type"`
	Stats    []string    `yaml:"stats,omitempty"`
	Matchers []Matcher `yaml:"matchers,omitempty"`
}

// Matcher is a flexible map to represent matcher configurations,
// e.g., {"isNumeric": true} or {"regex": "pattern"}.
type Matcher map[string]interface{}
