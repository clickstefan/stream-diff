# Contributing to Stream-Diff

Thank you for your interest in contributing to Stream-Diff! This document provides guidelines and information for contributors.

## 🚀 Getting Started

### Prerequisites

- Go 1.25 or later
- Git
- Make
- golangci-lint (installed via `make setup`)

### Development Setup

1. **Fork and clone the repository:**
   ```bash
   git clone https://github.com/yourusername/stream-diff.git
   cd stream-diff
   ```

2. **Set up development environment:**
   ```bash
   make setup    # Install development tools
   make deps     # Download dependencies
   ```

3. **Verify setup:**
   ```bash
   make quality-check  # Run all quality checks
   make test          # Run tests
   ```

## 🎯 Contribution Process

### 1. Find or Create an Issue

- Check existing [issues](https://github.com/clickstefan/stream-diff/issues)
- For bugs: Use the bug report template
- For features: Use the feature request template
- For questions: Use [Discussions](https://github.com/clickstefan/stream-diff/discussions)

### 2. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### 3. Make Changes

Follow our [coding standards](#coding-standards) and ensure:
- Code is well-documented
- Tests are included
- All quality checks pass

### 4. Test Your Changes

```bash
make pre-commit  # Run all checks
make test-integration  # Run integration tests
```

### 5. Submit Pull Request

- Use our [PR template](.github/pull_request_template.md)
- Reference related issues
- Include a clear description of changes
- Add screenshots for UI changes

## 📝 Coding Standards

### Go Code Style

We follow standard Go conventions plus additional rules:

1. **Formatting:** Use `gofmt` and `goimports`
2. **Linting:** All golangci-lint rules must pass
3. **Documentation:** All public functions must have godoc comments
4. **Error Handling:** Always wrap errors with context
5. **Testing:** Aim for >80% test coverage

### Code Structure

```go
// Package comment describing the package purpose
package example

import (
    // Standard library imports
    "context"
    "fmt"
    
    // Third-party imports
    "github.com/spf13/cobra"
    
    // Local imports
    "data-comparator/internal/pkg/config"
)

// Public functions need godoc comments
// ProcessData processes the input data and returns results.
// It returns an error if the data is invalid or processing fails.
func ProcessData(ctx context.Context, data []byte) (*Result, error) {
    if len(data) == 0 {
        return nil, fmt.Errorf("data cannot be empty")
    }
    
    // Implementation here
    return result, nil
}
```

### Error Handling

```go
// Good: Wrap errors with context
func readConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
    }
    
    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("failed to parse config %s: %w", path, err)
    }
    
    return &cfg, nil
}
```

### Testing

```go
func TestProcessData(t *testing.T) {
    tests := []struct {
        name        string
        input       []byte
        expected    *Result
        expectedErr string
    }{
        {
            name:     "valid input",
            input:    []byte("valid data"),
            expected: &Result{Status: "success"},
        },
        {
            name:        "empty input",
            input:       []byte{},
            expectedErr: "data cannot be empty",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            result, err := ProcessData(ctx, tt.input)
            
            if tt.expectedErr != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedErr)
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

## 🧪 Testing Guidelines

### Test Categories

1. **Unit Tests:** Test individual functions/methods
2. **Integration Tests:** Test component interactions
3. **End-to-End Tests:** Test complete workflows

### Test Structure

```bash
internal/pkg/config/
├── config.go           # Implementation
├── config_test.go      # Unit tests
└── testdata/          # Test fixtures
```

### Writing Good Tests

- **Use table-driven tests** for multiple scenarios
- **Test edge cases** and error conditions
- **Use meaningful test names** that describe the scenario
- **Include both positive and negative test cases**
- **Mock external dependencies** appropriately

### Running Tests

```bash
# Unit tests
make test

# With coverage
make test-coverage

# Integration tests
make test-integration

# Benchmarks
make benchmark
```

## 🔍 Code Review Process

### What We Look For

1. **Correctness:** Does the code do what it's supposed to?
2. **Performance:** Are there any obvious performance issues?
3. **Security:** Are there security vulnerabilities?
4. **Maintainability:** Is the code readable and well-structured?
5. **Testing:** Are there adequate tests?
6. **Documentation:** Is the code properly documented?

### Review Checklist

- [ ] Code follows our style guidelines
- [ ] All tests pass
- [ ] Test coverage is adequate
- [ ] Documentation is updated
- [ ] No security issues
- [ ] Performance considerations addressed
- [ ] Error handling is comprehensive
- [ ] Public APIs are well-documented

## 🏗️ Architecture Guidelines

### Package Organization

```
internal/pkg/
├── config/      # Configuration management
├── datareader/  # Data source abstractions
├── schema/      # Schema generation and analysis
├── compare/     # Comparison algorithms
└── report/      # Report generation
```

### Design Principles

1. **Single Responsibility:** Each package should have one clear purpose
2. **Dependency Injection:** Use interfaces for testability
3. **Error Handling:** Always provide context with errors
4. **Performance:** Consider memory and CPU efficiency
5. **Concurrency:** Use goroutines safely with proper synchronization

### Adding New Features

When adding a new feature:

1. **Start with interfaces** - Define the contract first
2. **Write tests first** - TDD approach preferred
3. **Keep it simple** - Avoid over-engineering
4. **Document everything** - Public APIs need godoc
5. **Consider backwards compatibility** - Don't break existing APIs

## 📊 Performance Guidelines

### Optimization Principles

1. **Measure first** - Use benchmarks to identify bottlenecks
2. **Memory efficiency** - Minimize allocations in hot paths
3. **Streaming processing** - Handle large datasets efficiently
4. **Concurrency** - Use goroutines for I/O bound operations
5. **Caching** - Cache expensive computations when appropriate

### Writing Benchmarks

```go
func BenchmarkProcessData(b *testing.B) {
    data := generateTestData(1000) // Setup
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ProcessData(context.Background(), data)
    }
}

func BenchmarkProcessDataParallel(b *testing.B) {
    data := generateTestData(1000)
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            ProcessData(context.Background(), data)
        }
    })
}
```

## 🐛 Bug Reports

### Before Reporting

1. Search existing issues
2. Try the latest version
3. Reduce to minimal reproduction case

### Bug Report Template

```markdown
**Description:**
Brief description of the bug

**Steps to Reproduce:**
1. Step one
2. Step two
3. See error

**Expected Behavior:**
What should have happened

**Actual Behavior:**
What actually happened

**Environment:**
- OS: [e.g., Ubuntu 22.04]
- Go version: [e.g., 1.25.0]
- Stream-diff version: [e.g., 1.0.0]

**Additional Context:**
Any other relevant information
```

## ✨ Feature Requests

### Feature Request Template

```markdown
**Is your feature request related to a problem?**
A clear description of the problem

**Describe the solution you'd like**
Clear description of the desired feature

**Describe alternatives you've considered**
Any alternative solutions or features considered

**Additional context**
Any other context or screenshots
```

## 🚀 Release Process

### Version Numbering

We follow [Semantic Versioning](https://semver.org/):
- `MAJOR.MINOR.PATCH`
- Major: Breaking changes
- Minor: New features (backwards compatible)
- Patch: Bug fixes (backwards compatible)

### Release Checklist

1. Update version in code
2. Update CHANGELOG.md
3. Ensure all tests pass
4. Create release PR
5. Tag release after merge
6. GitHub Actions handles the rest

## 💬 Communication

### Channels

- **Issues:** Bug reports and feature requests
- **Discussions:** General questions and ideas
- **PR Reviews:** Code-specific discussions
- **Email:** security@example.com for security issues

### Code of Conduct

We follow the [Contributor Covenant](https://www.contributor-covenant.org/):
- Be respectful and inclusive
- Welcome newcomers
- Focus on what's best for the project
- Show empathy towards others

## 🎖️ Recognition

Contributors will be recognized in:
- README.md contributors section
- Release notes
- Special contributor badges

Thank you for contributing to Stream-Diff! 🙏