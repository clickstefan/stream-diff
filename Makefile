# Data Stream Generator Makefile

# Build variables
BINARY_NAME=stream-generator
BUILD_DIR=bin
GO_FILES=$(shell find . -name "*.go")

# Default target
.PHONY: all
all: build

# Build the CLI tool
.PHONY: build
build: $(BUILD_DIR)/$(BINARY_NAME)

$(BUILD_DIR)/$(BINARY_NAME): $(GO_FILES)
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/generator/main.go

# Install dependencies
.PHONY: deps
deps:
	go mod download
	go mod tidy

# Run tests
.PHONY: test
test:
	go test -v ./...

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

# Development targets for testing different output formats
.PHONY: demo-csv
demo-csv: build
	./$(BUILD_DIR)/$(BINARY_NAME) -format csv -count 10 -header -schema examples/user_schema.yaml

.PHONY: demo-jsonl
demo-jsonl: build
	./$(BUILD_DIR)/$(BINARY_NAME) -format jsonl -count 10 -schema examples/schemas/kafka_events.yaml

.PHONY: demo-proto
demo-proto: build
	./$(BUILD_DIR)/$(BINARY_NAME) -format proto -count 5 -schema examples/schemas/iot_sensors.yaml

# Real-world examples
.PHONY: demo-ecommerce
demo-ecommerce: build
	@echo "Generating e-commerce order data..."
	./$(BUILD_DIR)/$(BINARY_NAME) -format csv -count 100 -schema examples/schemas/ecommerce_orders.yaml > /tmp/ecommerce_orders.csv
	@echo "Generated 100 e-commerce orders in /tmp/ecommerce_orders.csv"
	@head -5 /tmp/ecommerce_orders.csv

.PHONY: demo-logs
demo-logs: build
	@echo "Generating application log data..."
	./$(BUILD_DIR)/$(BINARY_NAME) -format jsonl -count 50 -rate 10 -schema examples/schemas/app_logs.yaml > /tmp/app_logs.jsonl
	@echo "Generated 50 log entries in /tmp/app_logs.jsonl"
	@head -3 /tmp/app_logs.jsonl

.PHONY: demo-kafka
demo-kafka: build
	@echo "Generating Kafka event stream data..."
	./$(BUILD_DIR)/$(BINARY_NAME) -format jsonl -count 25 -schema examples/schemas/kafka_events.yaml

.PHONY: demo-financial
demo-financial: build
	@echo "Generating financial transaction data..."
	./$(BUILD_DIR)/$(BINARY_NAME) -format csv -count 20 -schema examples/schemas/financial_transactions.yaml

# Performance testing
.PHONY: perf-test
perf-test: build
	@echo "Performance test: Generating 10,000 records at 1000/sec..."
	time ./$(BUILD_DIR)/$(BINARY_NAME) -format jsonl -count 10000 -rate 1000 > /dev/null
	@echo "Performance test completed"

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         - Build the CLI tool"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Install dependencies"
	@echo ""
	@echo "Demo targets:"
	@echo "  demo-csv      - Demo CSV output with user schema"
	@echo "  demo-jsonl    - Demo JSONL output with Kafka events"
	@echo "  demo-proto    - Demo protobuf output with IoT sensors"
	@echo "  demo-ecommerce- Generate e-commerce orders"
	@echo "  demo-logs     - Generate application logs"
	@echo "  demo-kafka    - Generate Kafka events"
	@echo "  demo-financial- Generate financial transactions"
	@echo "  perf-test     - Run performance test"