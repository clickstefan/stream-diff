package main

import (
	"data-comparator/internal/pkg/config"
	"data-comparator/internal/pkg/datareader"
	"data-comparator/internal/pkg/schema"
	"fmt"
	"io"
)

func main() {
	cfg, err := config.Load("testdata/testcase1_simple_csv/config1.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}
	cfg.Source.Path = cfg.Source.Path // path is already correct in the config file

	fmt.Printf("Config loaded: %+v\n", cfg)
	fmt.Printf("Source path: %s\n", cfg.Source.Path)

	reader, err := datareader.New(cfg.Source)
	if err != nil {
		fmt.Printf("Failed to create data reader: %v\n", err)
		return
	}
	defer reader.Close()

	fmt.Println("Data reader created successfully")

	// Let's manually sample a few records
	for i := 0; i < 3; i++ {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				fmt.Printf("EOF reached after %d records\n", i)
				break
			}
			fmt.Printf("Error reading record %d: %v\n", i, err)
			return
		}
		fmt.Printf("Record %d: %+v\n", i, record)

		// Test CollectFieldValues manually
		fieldValues := make(map[string][]interface{})
		schema.CollectFieldValues(record, fieldValues)
		fmt.Printf("Field values from record %d: %+v\n", i, fieldValues)
	}

	// Now test Generate
	// We need to create a new reader since we've already consumed records
	reader2, err := datareader.New(cfg.Source)
	if err != nil {
		fmt.Printf("Failed to create second data reader: %v\n", err)
		return
	}
	defer reader2.Close()

	generatedSchema, err := schema.Generate(reader2, cfg.Source.Sampler)
	if err != nil {
		fmt.Printf("Generate() error = %v\n", err)
		return
	}

	fmt.Printf("Generated schema: %+v\n", generatedSchema)
	if generatedSchema != nil {
		fmt.Printf("Number of fields: %d\n", len(generatedSchema.Fields))
		for name, field := range generatedSchema.Fields {
			fmt.Printf("Field %s: %+v\n", name, field)
		}
	}
}