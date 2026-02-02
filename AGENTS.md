# AGENTS.md - magicCommon Project Guidelines

This document provides guidelines for agentic coding agents working in the magicCommon Go project.

## Project Overview

magicCommon is a Go library providing common utilities, foundations, and frameworks for building applications. The project uses Go 1.24+ and follows standard Go conventions.

## Build Commands

### Testing
- **Run all tests**: `go test ./...`
- **Run tests with MySQL tag**: `go test -tags=mysql ./...`
- **Run tests for specific package**: `go test ./foundation/system`
- **Run single test**: `go test -run TestInvokeEntityFunc ./foundation/system`
- **Run tests with verbose output**: `go test -v ./...`
- **Run tests with coverage**: `go test -cover ./...`
- **Run benchmark tests**: `go test -bench=. ./...`

### Building
- **Build project**: `go build ./...`
- **Clean build cache**: `go clean -cache`
- **Install dependencies**: `go mod tidy`

### Linting & Formatting
- **Format code**: `gofmt -w .`
- **Check formatting**: `gofmt -d .`
- **Organize imports**: `goimports -w .`
- **Vet code**: `go vet ./...`

## Code Style Guidelines

### Package Structure
- Package names should be lowercase, single-word names
- Use subdirectories for logical separation (foundation/, framework/, etc.)
- Keep related functionality in the same package
- Use `internal/` for packages that should not be imported by external code

### Imports
- Group imports in this order:
  1. Standard library imports
  2. Third-party imports
  3. Local imports (from this module)
- Use blank lines between import groups
- Example:
```go
import (
    "fmt"
    "reflect"
    "testing"

    "github.com/stretchr/testify/assert"
    
    cd "github.com/muidea/magicCommon/def"
    "github.com/muidea/magicCommon/foundation/util"
)
```

### Naming Conventions
- **Variables**: Use camelCase (e.g., `entityVal`, `funcName`)
- **Constants**: Use CamelCase or UPPER_CASE for exported constants
- **Types**: Use PascalCase (e.g., `MockEntity`, `KVCache`)
- **Interfaces**: Use PascalCase, often ending with "er" (e.g., `Cache`, `Walker`)
- **Methods**: Use PascalCase for exported methods
- **Private members**: Start with lowercase
- **Acronyms**: Keep as all caps (e.g., `ID`, `URL`, `HTTP`)

### Error Handling
- Use the custom error type from `github.com/muidea/magicCommon/def` package
- Error variables should be named `err` or `errVal`
- Return errors as `*cd.Error` type
- Use `cd.NewError()` to create new errors with error codes
- Handle panics with recover in critical functions
- Example error handling pattern:
```go
func SomeFunction() (err *cd.Error) {
    defer func() {
        if errInfo := recover(); errInfo != nil {
            err = cd.NewError(cd.Unexpected, fmt.Sprintf("recover! %v", errInfo))
        }
    }()
    
    // Function logic
    return nil
}
```

### Types and Interfaces
- Define interfaces for abstraction
- Use generics where appropriate (Go 1.24+)
- Prefer composition over inheritance
- Use type aliases for clarity when needed
- Example interface pattern:
```go
type Cache interface {
    Put(key string, val interface{}) *cd.Error
    Fetch(key string) (interface{}, *cd.Error)
    Remove(key string) *cd.Error
}
```

### Testing
- Use `testing` package from standard library
- Use `github.com/stretchr/testify/assert` for assertions
- Test files should be named `*_test.go`
- Test functions should start with `Test` prefix
- Use table-driven tests for multiple test cases
- Use subtests with `t.Run()` for better organization
- Example test structure:
```go
func TestFunctionName(t *testing.T) {
    t.Run("Test case description", func(t *testing.T) {
        // Setup
        // Execution
        // Assertion
        assert.Nil(t, result)
        assert.Equal(t, expected, actual)
    })
}
```

### Benchmarking
- Benchmark files should be named `*_benchmark_test.go`
- Benchmark functions should start with `Benchmark` prefix
- Use `b.ResetTimer()` and `b.StopTimer()` appropriately
- Example benchmark:
```go
func BenchmarkCache_Put(b *testing.B) {
    cache := NewCache(nil)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        cache.Put(fmt.Sprintf("key%d", i), i)
    }
}
```

### Logging
- Use `github.com/muidea/seelog` for logging
- Log levels: DEBUG, INFO, WARN, ERROR
- Include contextual information in log messages
- Example logging:
```go
import "github.com/muidea/seelog"

seelog.Warnf("illegal value, not string, value:%v", value)
```

### Documentation
- Document exported functions, types, and packages
- Use GoDoc style comments
- Keep comments concise and focused on "why" not "what"
- Example:
```go
// InvokeEntityFunc invokes a method by name on an entity value.
// It handles parameter conversion and error recovery.
func InvokeEntityFunc(entityVal interface{}, funcName string, params ...interface{}) (err *cd.Error) {
```

### Concurrency
- Use `golang.org/x/sync` for synchronization primitives
- Consider thread safety for shared resources
- Use channels for communication between goroutines
- Document concurrent access patterns

### Configuration
- Use TOML format for configuration files
- Configuration files are in `config/` directories
- Use `github.com/pelletier/go-toml/v2` for parsing

## Project Structure

```
magicCommon/
├── def/              # Common definitions and error types
├── foundation/       # Core utilities and foundations
│   ├── cache/       # Caching implementations
│   ├── dao/         # Data access objects
│   ├── path/        # Path utilities
│   ├── system/      # System utilities
│   └── util/        # General utilities
├── framework/       # Framework components
│   └── configuration/ # Configuration framework
├── event/           # Event system
├── monitoring/      # Monitoring utilities
├── session/         # Session management
├── task/            # Task management
└── test/            # Test utilities
```

## Database Support

- MySQL: Use `-tags=mysql` for MySQL-specific tests
- PostgreSQL: Default database for tests
- Both databases are tested in CI/CD pipeline

## CI/CD Pipeline

- Tests run on push to master, feature/*, and bugfix/* branches
- MySQL and PostgreSQL services are provisioned for testing
- Go 1.24 is used for builds and tests

## Best Practices

1. **Error First**: Always check and handle errors immediately
2. **Defensive Programming**: Validate inputs, handle edge cases
3. **Immutable by Default**: Use const where possible, avoid mutation
4. **Single Responsibility**: Functions should do one thing well
5. **DRY (Don't Repeat Yourself)**: Extract common patterns
6. **KISS (Keep It Simple)**: Prefer simple solutions over complex ones
7. **YAGNI (You Ain't Gonna Need It)**: Don't add features until needed

## Common Patterns

- Use factory functions for object creation (e.g., `NewCache()`)
- Use builder pattern for complex object construction
- Use dependency injection for testability
- Use context for request-scoped values and cancellation

## When Making Changes

1. Run tests before and after changes
2. Ensure all tests pass, including MySQL-tagged tests
3. Format code with `gofmt`
4. Update documentation for new/changed functionality
5. Consider backward compatibility for public APIs
6. Add tests for new functionality
7. Update benchmarks if performance-critical code changes