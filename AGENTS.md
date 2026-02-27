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
- **Run tests with count flag**: `go test -count=1 ./...` (disable test caching)
- **Run parallel benchmark tests**: `go test -bench=. -benchtime=5s ./...`

### Building
- **Build project**: `go build ./...`
- **Clean build cache**: `go clean -cache`
- **Install dependencies**: `go mod tidy`

### Makefile Commands
- **Run all (build, test, lint)**: `make all`
- **Run tests**: `make test`
- **Run tests with coverage**: `make test-coverage`
- **Run lint checks**: `make lint`
- **Format code**: `make fmt`
- **Check formatting**: `make fmt-check`
- **Run vet**: `make vet`
- **Clean build files**: `make clean`
- **Install dependencies**: `make deps`
- **Install dev tools**: `make dev-tools`
- **Run benchmarks**: `make bench`
- **Run tests for specific package**: `make test-foundation/system`
- **Run benchmarks for specific package**: `make bench-foundation/cache`

### Linting & Formatting
- **Format code**: `gofmt -w .` or `make fmt`
- **Check formatting**: `gofmt -d .` or `make fmt-check`
- **Organize imports**: `goimports -w .`
- **Vet code**: `go vet ./...` or `make vet`
- **Run comprehensive lint**: `make lint` (includes vet, format check, and optional golangci-lint)
- **Security scan**: `gosec ./...` (requires installation: `go install github.com/securego/gosec/v2/cmd/gosec@latest`)

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
- Always check and handle errors immediately

### Types and Interfaces
- Define interfaces for abstraction
- Use generics where appropriate (Go 1.24+)
- Prefer composition over inheritance
- Use type aliases for clarity when needed

### Testing
- Use `testing` package from standard library
- Use `github.com/stretchr/testify/assert` for assertions
- Test files should be named `*_test.go`
- Test functions should start with `Test` prefix
- Use table-driven tests for multiple test cases
- Use subtests with `t.Run()` for better organization
- Example test structure (from foundation/system/system_test.go:28):
```go
// TestInvokeEntityFuncNoMethod tests the scenario where the method does not exist on the entityVal
func TestInvokeEntityFuncNoMethod(t *testing.T) {
    entityVal := &MockEntity{}
    funcName := "NonExistentMethod"

    result := InvokeEntityFunc(entityVal, funcName)
    assert.NotNil(t, result)
}
```
- For database tests, use `-tags=mysql` flag for MySQL-specific tests
- Use `t.Parallel()` for tests that can run concurrently when appropriate

### Benchmarking
- Benchmark files should be named `*_benchmark_test.go`
- Benchmark functions should start with `Benchmark` prefix
- Use `b.ResetTimer()` and `b.StopTimer()` appropriately
- Use `b.RunParallel()` for concurrent benchmarks
- Run parallel benchmark tests: `go test -bench=. -benchtime=5s ./...`

### Logging
- Use `github.com/muidea/seelog` for logging
- Log levels: DEBUG, INFO, WARN, ERROR
- Include contextual information in log messages

### Documentation
- Document exported functions, types, and packages
- Use GoDoc style comments
- Keep comments concise and focused on "why" not "what"
- Example (from foundation/system/system.go:11):
```go
// InvokeEntityFunc invokes a method by name on an entity value.
// It handles parameter conversion and error recovery.
func InvokeEntityFunc(entityVal interface{}, funcName string, params ...interface{}) (err *cd.Error) {
```



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
│   ├── application/ # Application framework
│   ├── configuration/ # Configuration framework
│   ├── plugin/      # Plugin system
│   └── service/     # Service framework
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
- CI workflow: `.github/workflows/ci.yml`
- MySQL test database: `testdb` with root password `rootkit`
- PostgreSQL test database: `testdb` with user `postgres` and password `rootkit`

## Best Practices

1. **Error First**: Always check and handle errors immediately
2. **Defensive Programming**: Validate inputs, handle edge cases
3. **Immutable by Default**: Use const where possible, avoid mutation
4. **Single Responsibility**: Functions should do one thing well
5. **DRY (Don't Repeat Yourself)**: Extract common patterns
6. **KISS (Keep It Simple)**: Prefer simple solutions over complex ones
7. **YAGNI (You Ain't Gonna Need It)**: Don't add features until needed