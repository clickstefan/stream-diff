# Data Stream Generator CLI

A powerful command-line tool for generating realistic test data streams for databases, Kafka topics, logs, and other data systems. The generator creates data based on schema definitions and supports multiple output formats with proper delimiters for easy piping to other tools.

## Features

- **Multiple Output Formats**: CSV, JSONL, and protobuf-style JSON
- **Schema-Based Generation**: Uses YAML schema files to define data structure and types
- **Real-World Data Patterns**: Comprehensive patterns for e-commerce, financial, IoT, logging, and more
- **Flexible Output**: Streams to stdout with configurable delimiters
- **Rate Limiting**: Control generation speed for performance testing
- **Backpressure Handling**: Memory-efficient streaming with proper resource management
- **Reproducible Output**: Seed-based random generation for consistent testing

## Installation

```bash
make build
```

This creates the `bin/stream-generator` executable.

## Usage

```bash
./bin/stream-generator [options]

Options:
  -schema string
        Path to schema YAML file (optional)
  -format string
        Output format: csv, jsonl, proto (default "jsonl")
  -count int
        Maximum number of records to generate (0 = unlimited) (default 100)
  -rate int
        Records per second (0 = unlimited) (default 0)
  -buffer int
        Buffer size for backpressure handling (default 100)
  -seed int
        Random seed for reproducible output (0 = use current time) (default 0)
  -delimiter string
        Custom delimiter (default: \n for csv/jsonl, \n for proto)
  -header
        Include CSV header row (CSV format only) (default true)
```

## Examples

### Basic Usage

Generate 100 records in JSONL format:
```bash
./bin/stream-generator -count 100
```

Generate CSV with headers:
```bash
./bin/stream-generator -format csv -count 1000 -header > data.csv
```

Generate protobuf-style JSON:
```bash
./bin/stream-generator -format proto -count 500
```

### Real-World Scenarios

**E-commerce Orders:**
```bash
./bin/stream-generator -schema examples/schemas/ecommerce_orders.yaml -format csv -count 10000 > orders.csv
```

**Kafka Event Stream:**
```bash
./bin/stream-generator -schema examples/schemas/kafka_events.yaml -format jsonl -rate 1000 | kafka-console-producer.sh --topic events
```

**Application Logs:**
```bash
./bin/stream-generator -schema examples/schemas/app_logs.yaml -format jsonl -count 50000 > app.log
```

**Financial Transactions:**
```bash
./bin/stream-generator -schema examples/schemas/financial_transactions.yaml -format csv -count 1000000 > transactions.csv
```

**IoT Sensor Data:**
```bash
./bin/stream-generator -schema examples/schemas/iot_sensors.yaml -format jsonl -rate 100 -count 0 | mqtt-publisher
```

### Performance Testing

Generate high-throughput data:
```bash
# Generate 1M records at 10k/sec
./bin/stream-generator -count 1000000 -rate 10000 | wc -l

# Continuous generation for load testing
./bin/stream-generator -count 0 -rate 1000 | your-consumer-app
```

## Schema Files

Schema files define the structure and types of generated data. The generator includes several real-world schema examples:

- `examples/schemas/ecommerce_orders.yaml` - E-commerce order data
- `examples/schemas/kafka_events.yaml` - User activity events
- `examples/schemas/app_logs.yaml` - Application log entries
- `examples/schemas/iot_sensors.yaml` - IoT sensor readings
- `examples/schemas/financial_transactions.yaml` - Banking/payment transactions

### Schema Format

```yaml
key: field_name  # Primary key field
max_key_size: 10 # Maximum key length
fields:
  field_name:
    type: string|numeric|datetime|boolean|object|array
    stats: ["cardinality", "availability", "min", "max", "avg"]
```

## Data Patterns

The generator automatically creates realistic data based on field names and includes comprehensive patterns for:

### Business Data
- **E-commerce**: Orders, products, customers, payments
- **Financial**: Transactions, accounts, currencies, risk scores
- **CRM**: Users, contacts, interactions, sales data

### Technical Data
- **Logging**: Log levels, error codes, response times, stack traces
- **Web Analytics**: Page views, clicks, sessions, user agents
- **System Metrics**: CPU, memory, network, performance data

### IoT & Sensors
- **Environmental**: Temperature, humidity, pressure, air quality
- **Device Management**: Battery levels, firmware versions, connectivity
- **Location Data**: GPS coordinates, addresses, time zones

### Formats & Identifiers
- **IDs**: UUIDs, sequential IDs, custom formats
- **Network**: IP addresses, MAC addresses, URLs
- **Contact**: Emails, phone numbers, addresses

## Output Formats

### CSV
```csv
user_id,email,age,city,plan_type,last_login
1,user1@example.com,42,New York,premium,2025-09-15T12:00:00Z
2,user2@example.com,28,Los Angeles,basic,2025-09-15T11:30:00Z
```

### JSONL (JSON Lines)
```jsonl
{"user_id":1,"email":"user1@example.com","age":42,"city":"New York","plan_type":"premium","last_login":"2025-09-15T12:00:00Z"}
{"user_id":2,"email":"user2@example.com","age":28,"city":"Los Angeles","plan_type":"basic","last_login":"2025-09-15T11:30:00Z"}
```

### Protobuf-style JSON
```json
{"user_id":1,"email":"user1@example.com","age":42,"city":"New York","plan_type":"premium","last_login":"2025-09-15T12:00:00Z"}
{"user_id":2,"email":"user2@example.com","age":28,"city":"Los Angeles","plan_type":"basic","last_login":"2025-09-15T11:30:00Z"}
```

## Make Targets

Convenient make targets are available for common tasks:

```bash
make build          # Build the CLI tool
make test           # Run tests
make clean          # Clean build artifacts

# Demo commands
make demo-csv       # Demo CSV output
make demo-jsonl     # Demo JSONL output
make demo-proto     # Demo protobuf output
make demo-ecommerce # Generate e-commerce data
make demo-logs      # Generate application logs
make demo-kafka     # Generate Kafka events
make demo-financial # Generate financial data
make perf-test      # Performance testing
```

## Integration Examples

### With Kafka
```bash
# Stream events to Kafka topic
./bin/stream-generator -schema kafka_events.yaml -rate 1000 -count 0 | \
  kafka-console-producer.sh --bootstrap-server localhost:9092 --topic user-events
```

### With Database Import
```bash
# Generate CSV for database import
./bin/stream-generator -schema ecommerce_orders.yaml -format csv -count 1000000 | \
  psql -c "COPY orders FROM STDIN CSV HEADER"
```

### With Log Analysis Tools
```bash
# Generate logs for testing log parsers
./bin/stream-generator -schema app_logs.yaml -count 100000 | \
  logstash -f logstash.conf
```

### With Load Testing
```bash
# Generate realistic API payloads
./bin/stream-generator -schema api_requests.yaml -rate 500 | \
  while read line; do curl -X POST -d "$line" http://api.example.com/endpoint; done
```

## Performance Characteristics

- **Memory Efficient**: Constant memory usage regardless of generation rate
- **High Throughput**: Tested at 10,000+ records/second
- **Backpressure Handling**: Automatically slows when consumers can't keep up
- **Resource Management**: Proper cleanup and graceful shutdown

## Contributing

The generator is designed to be easily extensible:

1. Add new data patterns in `cmd/generator/main.go`
2. Create new schema examples in `examples/schemas/`
3. Extend format support in the output functions
4. Add new field type handlers in the generator package