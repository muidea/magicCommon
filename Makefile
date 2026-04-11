# MagicCommon 项目 Makefile

.PHONY: all build test lint vet fmt clean help

# 默认目标
all: build test lint

# 构建项目
build:
	@echo "Building project..."
	go build ./...

# 运行测试
test:
	@echo "Running tests..."
	go test ./... --count=1

# 运行测试并生成覆盖率报告
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 运行 lint 检查
lint:
	@echo "Running code quality checks..."
	@echo "1. Running go vet..."
	@go vet ./... 2>&1 | (grep -v "^vendor/" || true)
	@echo "2. Checking code format (excluding vendor)..."
	@if find . -name "*.go" -not -path "./vendor/*" -exec gofmt -d {} + 2>/dev/null | grep -q '^'; then \
		echo "Code is not formatted correctly. Run 'make fmt' to fix."; \
		find . -name "*.go" -not -path "./vendor/*" -exec gofmt -d {} + 2>/dev/null | head -50; \
		exit 1; \
	else \
		echo "Code is properly formatted."; \
	fi
	@echo "3. Trying golangci-lint (optional)..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found, skipping..."; \
		echo "To install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 运行 vet 检查
vet:
	@echo "Running go vet..."
	go vet ./...

# 格式化代码
fmt:
	@echo "Formatting code..."
	@find . -name "*.go" -not -path "./vendor/*" -exec gofmt -w {} +
	@if command -v goimports >/dev/null 2>&1; then \
		find . -name "*.go" -not -path "./vendor/*" -exec goimports -w {} +; \
		echo "goimports completed."; \
	else \
		echo "goimports not found, skipping..."; \
		echo "To install: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

# 检查代码格式
fmt-check:
	@echo "Checking code format (excluding vendor)..."
	@if find . -name "*.go" -not -path "./vendor/*" -exec gofmt -d {} + 2>/dev/null | grep -q '^'; then \
		echo "Code is not formatted correctly. Run 'make fmt' to fix."; \
		find . -name "*.go" -not -path "./vendor/*" -exec gofmt -d {} + 2>/dev/null | head -50; \
		exit 1; \
	else \
		echo "Code is properly formatted."; \
	fi

# 清理构建文件
clean:
	@echo "Cleaning up..."
	go clean -cache
	rm -f coverage.out coverage.html

# 安装依赖
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# 安装开发工具
dev-tools:
	@echo "Installing development tools..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/stretchr/testify/assert@latest

# 运行基准测试
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchtime=3s ./...

# 运行特定包的测试
test-%:
	@echo "Running tests for package $*..."
	go test ./$*

# 运行特定包的基准测试
bench-%:
	@echo "Running benchmarks for package $*..."
	go test -bench=. -benchtime=3s ./$*

# 显示帮助
help:
	@echo "Available targets:"
	@echo "  all          - Build, test and lint (default)"
	@echo "  build        - Build the project"
	@echo "  test         - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint         - Run golangci-lint"
	@echo "  vet          - Run go vet"
	@echo "  fmt          - Format code"
	@echo "  fmt-check    - Check code format"
	@echo "  clean        - Clean build files"
	@echo "  deps         - Install dependencies"
	@echo "  dev-tools    - Install development tools"
	@echo "  bench        - Run benchmarks"
	@echo "  test-<pkg>   - Run tests for specific package"
	@echo "  bench-<pkg>  - Run benchmarks for specific package"
	@echo "  help         - Show this help message"
