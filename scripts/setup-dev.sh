#!/bin/bash

# magicCommon 开发环境设置脚本

set -e

echo "=== magicCommon 开发环境设置 ==="

# 检查 Go 版本
echo "1. 检查 Go 版本..."
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.24"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "错误: 需要 Go $REQUIRED_VERSION 或更高版本，当前版本: $GO_VERSION"
    exit 1
fi
echo "✓ Go 版本符合要求: $GO_VERSION"

# 安装开发工具
echo "2. 安装开发工具..."
echo "   安装 goimports..."
go install golang.org/x/tools/cmd/goimports@latest
echo "   安装 golangci-lint..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
echo "   安装 gosec..."
go install github.com/securego/gosec/v2/cmd/gosec@latest
echo "✓ 开发工具安装完成"

# 安装项目依赖
echo "3. 安装项目依赖..."
go mod tidy
go mod download
echo "✓ 项目依赖安装完成"

# 检查 Docker 是否可用（用于测试数据库）
echo "4. 检查测试环境..."
if command -v docker &> /dev/null; then
    echo "   Docker 已安装"
    
    # 检查 MySQL 容器
    if ! docker ps -a --format '{{.Names}}' | grep -q 'mysql-test'; then
        echo "   提示: 可以使用以下命令启动 MySQL 测试容器:"
        echo "   docker run -d --name mysql-test \\"
        echo "     -e MYSQL_DATABASE=testdb \\"
        echo "     -e MYSQL_ROOT_PASSWORD=rootkit \\"
        echo "     -p 3306:3306 \\"
        echo "     mysql:5.7"
    else
        echo "   ✓ MySQL 测试容器已存在"
    fi
    
    # 检查 PostgreSQL 容器
    if ! docker ps -a --format '{{.Names}}' | grep -q 'postgres-test'; then
        echo "   提示: 可以使用以下命令启动 PostgreSQL 测试容器:"
        echo "   docker run -d --name postgres-test \\"
        echo "     -e POSTGRES_USER=postgres \\"
        echo "     -e POSTGRES_PASSWORD=rootkit \\"
        echo "     -e POSTGRES_DB=testdb \\"
        echo "     -p 5432:5432 \\"
        echo "     postgres:17.2-alpine"
    else
        echo "   ✓ PostgreSQL 测试容器已存在"
    fi
else
    echo "   警告: Docker 未安装，无法运行数据库相关测试"
    echo "   提示: 请安装 Docker 或手动配置测试数据库"
fi

# 运行初始验证
echo "5. 运行初始验证..."
echo "   构建项目..."
if go build ./...; then
    echo "   ✓ 构建成功"
else
    echo "   ✗ 构建失败"
    exit 1
fi

echo "   检查代码格式..."
if make fmt-check > /dev/null 2>&1; then
    echo "   ✓ 代码格式正确"
else
    echo "   ✗ 代码格式有问题，运行 'make fmt' 修复"
    make fmt-check | head -20
fi

echo "   运行静态分析..."
if go vet ./... 2>&1 | grep -v "^vendor/"; then
    echo "   ✗ 静态分析发现问题"
else
    echo "   ✓ 静态分析通过"
fi

# 创建 git hooks
echo "6. 设置 Git hooks..."
HOOKS_DIR=".git/hooks"
PRE_COMMIT_HOOK="$HOOKS_DIR/pre-commit"

if [ -d ".git" ]; then
    cat > "$PRE_COMMIT_HOOK" << 'EOF'
#!/bin/bash

echo "=== 运行提交前检查 ==="

# 检查代码格式
echo "1. 检查代码格式..."
if make fmt-check > /dev/null 2>&1; then
    echo "   ✓ 代码格式正确"
else
    echo "   ✗ 代码格式有问题，请运行 'make fmt' 修复"
    make fmt-check | head -20
    exit 1
fi

# 运行静态分析
echo "2. 运行静态分析..."
if go vet ./... 2>&1 | grep -v "^vendor/"; then
    echo "   ✗ 静态分析发现问题"
    exit 1
else
    echo "   ✓ 静态分析通过"
fi

# 运行测试
echo "3. 运行测试..."
if go test ./... 2>&1 | tail -1 | grep -q "ok"; then
    echo "   ✓ 测试通过"
else
    echo "   ✗ 测试失败"
    exit 1
fi

echo "=== 所有检查通过，可以提交 ==="
EOF
    
    chmod +x "$PRE_COMMIT_HOOK"
    echo "✓ Git pre-commit hook 已设置"
else
    echo "   提示: 当前目录不是 Git 仓库，跳过 Git hooks 设置"
fi

echo ""
echo "=== 开发环境设置完成 ==="
echo ""
echo "常用命令:"
echo "  make all      - 构建、测试、代码检查"
echo "  make build    - 构建项目"
echo "  make test     - 运行测试"
echo "  make lint     - 代码质量检查"
echo "  make vet      - 静态分析"
echo "  make fmt      - 格式化代码"
echo "  make fmt-check - 检查代码格式"
echo "  make clean    - 清理构建文件"
echo ""
echo "提交代码前请确保运行 'make all' 通过所有检查！"