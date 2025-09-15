package main

import (
	"data-comparator/internal/pkg/config"
	"data-comparator/internal/pkg/datareader"
	"fmt"
)

type workItem struct {
	data   interface{}
	prefix string
}

func CollectFieldValuesDebug(data interface{}, fieldValues map[string][]interface{}) {
	queue := []workItem{{data: data, prefix: ""}}
	fmt.Printf("Initial queue: %+v\n", queue)

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]
		
		fmt.Printf("Processing item: data=%+v, prefix='%s'\n", item.data, item.prefix)

		if item.data == nil {
			fmt.Println("  -> Skipping nil data")
			continue
		}

		if m, ok := item.data.(map[string]interface{}); ok {
			fmt.Printf("  -> Processing as map with %d keys\n", len(m))
			if item.prefix != "" {
				fieldValues[item.prefix] = append(fieldValues[item.prefix], m)
				fmt.Printf("  -> Added map to fieldValues[%s]\n", item.prefix)
			} else {
				fmt.Println("  -> Not adding map (prefix is empty)")
			}
			for key, value := range m {
				newKey := key
				if item.prefix != "" {
					newKey = item.prefix + "." + key
				}
				queue = append(queue, workItem{data: value, prefix: newKey})
				fmt.Printf("  -> Added to queue: data=%+v, prefix='%s'\n", value, newKey)
			}
		} else if a, ok := item.data.([]interface{}); ok {
			fmt.Printf("  -> Processing as array with %d elements\n", len(a))
			if item.prefix != "" {
				fieldValues[item.prefix] = append(fieldValues[item.prefix], a)
				fmt.Printf("  -> Added array to fieldValues[%s]\n", item.prefix)
			}
			arrayKey := item.prefix + "[]"
			for _, v := range a {
				queue = append(queue, workItem{data: v, prefix: arrayKey})
			}
		} else {
			fmt.Printf("  -> Processing as simple value: %+v\n", item.data)
			if item.prefix != "" {
				fieldValues[item.prefix] = append(fieldValues[item.prefix], item.data)
				fmt.Printf("  -> Added value to fieldValues[%s]\n", item.prefix)
			} else {
				fmt.Println("  -> Not adding value (prefix is empty)")
			}
		}
		
		fmt.Printf("  -> Current fieldValues: %+v\n", fieldValues)
		fmt.Printf("  -> Remaining queue length: %d\n", len(queue))
		fmt.Println()
	}
}

func main() {
	cfg, err := config.Load("testdata/testcase1_simple_csv/config1.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	reader, err := datareader.New(cfg.Source)
	if err != nil {
		fmt.Printf("Failed to create data reader: %v\n", err)
		return
	}
	defer reader.Close()

	record, err := reader.Read()
	if err != nil {
		fmt.Printf("Failed to read record: %v\n", err)
		return
	}

	fmt.Printf("Record: %+v\n", record)
	fmt.Printf("Record type: %T\n", record)
	
	fieldValues := make(map[string][]interface{})
	CollectFieldValuesDebug(record, fieldValues)
	
	fmt.Printf("Final field values: %+v\n", fieldValues)
}