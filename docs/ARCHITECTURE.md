# Stream-Diff Architecture Documentation

## Overview

Stream-Diff is designed as a modern, extensible data comparison tool built with Go 1.25+ and following current best practices for CLI applications, testing, and security.

## System Architecture

### High-Level Components

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   CLI Layer     │    │   Core Engine    │    │  Data Sources   │
│                 │    │                  │    │                 │
│ • Cobra Commands│    │ • Schema Gen     │    │ • CSV Reader    │
│ • AI Suggestions│    │ • Comparison     │    │ • JSON Reader   │
│ • Validation    │    │ • Report Gen     │    │ • Future: API   │
│ • Logging       │    │ • Statistics     │    │ • Future: DB    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────┴─────────────┐
                    │   Configuration Layer    │
                    │                           │
                    │ • YAML Config Files       │
                    │ • Viper Integration       │
                    │ • Environment Variables   │
                    │ • Command Line Flags      │
                    └───────────────────────────┘
```

### Component Details

#### 1. CLI Layer (`cmd/`)
- **Root Command**: Main application entry point with global configuration
- **Compare Command**: Data comparison orchestration with AI insights
- **Validate Command**: Configuration validation with smart recommendations
- **Version Command**: Build and version information display

**Design Decisions:**
- Uses Cobra framework for consistent CLI UX
- Structured logging with zerolog for observability
- AI-powered help system with contextual suggestions
- Comprehensive error handling with actionable messages

#### 2. Core Engine (`internal/pkg/`)

**Configuration Package (`config/`)**
```go
type Config struct {
    Source Source `yaml:"source"`
}

type Source struct {
    Type         string        `yaml:"type"`
    Path         string        `yaml:"path"`
    ParserConfig *ParserConfig `yaml:"parser_config,omitempty"`
    Sampler      *Sampler      `yaml:"sampler,omitempty"`
}
```

**Data Reader Package (`datareader/`)**
```go
type DataReader interface {
    Read() (Record, error)
    Close() error
}
```

**Schema Package (`schema/`)**
```go
type Schema struct {
    Key        string           `yaml:"key"`
    MaxKeySize int              `yaml:"max_key_size,omitempty"`
    Fields     map[string]*Field `yaml:"fields"`
}
```

### Data Flow

1. **Configuration Loading**
   ```
   YAML Config → Viper → Config Struct → Validation → Ready for Use
   ```

2. **Schema Generation**
   ```
   Data Source → DataReader → Sample Records → Field Analysis → Schema Object
   ```

3. **Comparison Process**
   ```
   Schema1 + Schema2 → Compatibility Check → AI Insights → Report Generation
   ```

4. **AI Enhancement Layer**
   ```
   Raw Results → Pattern Analysis → Insight Generation → User-Friendly Output
   ```

## Key Design Patterns

### 1. Strategy Pattern (Data Readers)
Different data source types implement the same `DataReader` interface:
- CSV files
- JSON-Lines files  
- Future: Database connections, APIs, Kafka streams

### 2. Factory Pattern (Reader Creation)
```go
func New(cfg config.Source) (DataReader, error) {
    switch cfg.Type {
    case "csv":
        return NewCSVReader(cfg)
    case "json":
        return NewJSONReader(cfg)
    default:
        return nil, fmt.Errorf("unsupported source type: %s", cfg.Type)
    }
}
```

### 3. Command Pattern (CLI Commands)
Each CLI command is self-contained with its own flags, validation, and execution logic.

### 4. Observer Pattern (AI Insights)
The comparison engine generates events that the AI insight system observes to provide recommendations.

## Performance Considerations

### Memory Management
- **Streaming Processing**: Large files processed in chunks
- **Memory Pools**: Reuse objects to reduce GC pressure
- **Sampling**: Configurable sampling for large datasets
- **Bounded Channels**: Control memory usage in concurrent operations

### Concurrency
- **Schema Generation**: Concurrent field analysis
- **I/O Operations**: Non-blocking file operations where possible
- **CPU-Bound Tasks**: Worker pool pattern for comparison operations

### Optimization Features
- **Early Termination**: Stop processing when key mismatches detected
- **Caching**: Cache expensive computations (regex matching, type detection)
- **Lazy Loading**: Load data only when needed

## Security Architecture

### Input Validation
```go
func validateConfig(cfg *config.Config, configPath string) error {
    // File existence checks
    // Path traversal protection  
    // Size limits
    // Type validation
}
```

### Supply Chain Security
- **SBOM Generation**: Track all dependencies
- **Vulnerability Scanning**: Regular security updates
- **Dependency Pinning**: Controlled dependency upgrades
- **Code Signing**: Verify build artifacts

### Runtime Security
- **File Permissions**: Minimal required permissions
- **Memory Safety**: Go's memory safety + static analysis
- **Input Sanitization**: All user inputs validated
- **Error Information**: No sensitive data in error messages

## Testing Strategy

### Test Pyramid

```
                    ┌─────────────────┐
                    │   End-to-End    │
                    │    Tests        │
                    │   (Integration) │
                    └─────────────────┘
                  ┌─────────────────────┐
                  │   Integration Tests │
                  │  (Component Tests)  │
                  └─────────────────────┘
               ┌──────────────────────────┐
               │       Unit Tests         │
               │   (Fast & Isolated)     │
               └──────────────────────────┘
```

**Unit Tests (70%)**
- Individual function testing
- Mock dependencies
- Fast execution
- High coverage

**Integration Tests (20%)**
- Component interaction testing
- Real file system usage
- Database connections (future)
- API interactions (future)

**End-to-End Tests (10%)**
- Full CLI workflow testing
- Real data processing
- Cross-platform validation
- Performance regression detection

## AI Integration Points

### 1. Configuration Validation
- Pattern recognition for common configuration errors
- Performance optimization suggestions
- Compatibility warnings between sources

### 2. Schema Analysis
- Field type mismatch detection
- Missing field identification
- Data quality issue prediction
- Migration pathway suggestions

### 3. Error Handling
- Context-aware error messages
- Suggested solutions based on error patterns
- Learning from user interactions
- Proactive issue prevention

### 4. Performance Optimization
- Automatic sampling size recommendations
- Memory usage optimization suggestions
- Query optimization for large datasets
- Resource usage monitoring

## Extensibility Points

### New Data Source Types
```go
// Implement DataReader interface
type NewSourceReader struct {
    // source-specific fields
}

func (r *NewSourceReader) Read() (Record, error) {
    // implementation
}

func (r *NewSourceReader) Close() error {
    // cleanup
}
```

### New Output Formats
```go
// Add to comparison output handling
func outputResult(result ComparisonResult, format string) error {
    switch format {
    case "yaml":
        return outputYAML(result)
    case "json":
        return outputJSON(result)  
    case "xml": // New format
        return outputXML(result)
    }
}
```

### New AI Insights
```go
// Add to insight generation
func generateInsights(schema1, schema2 *Schema) []AIInsight {
    var insights []AIInsight
    
    // Existing insights...
    
    // New insight type
    insights = append(insights, detectDataQualityIssues(schema1, schema2)...)
    
    return insights
}
```

## Future Enhancements

### Planned Features
1. **Advanced Data Sources**
   - Database connections (PostgreSQL, MySQL, MongoDB)
   - REST API endpoints
   - Message queues (Kafka, RabbitMQ)
   - Cloud storage (S3, GCS, Azure Blob)

2. **Enhanced AI Capabilities**
   - Machine learning-based pattern detection
   - Automated data mapping suggestions
   - Predictive data quality scoring
   - Natural language query interface

3. **Performance Improvements**
   - Distributed processing support
   - GPU acceleration for large datasets
   - Advanced caching mechanisms
   - Real-time streaming comparisons

4. **Enterprise Features**
   - Role-based access control
   - Audit logging
   - Integration with data catalogs
   - Compliance reporting

### Architecture Evolution
The current architecture supports these future enhancements through:
- Interface-based design for easy extension
- Plugin architecture preparation
- Modular component organization
- Comprehensive testing framework