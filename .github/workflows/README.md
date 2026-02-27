# CI/CD 工作流说明

本项目使用 GitHub Actions 进行持续集成和持续部署。以下是配置的工作流说明：

## 工作流文件

### 1. `ci.yml` - 持续集成流水线
**触发条件**:
- `push` 到以下分支: `master`, `feature/*`, `bugfix/*`, `hotfix/*`
- `pull_request` 到以下分支: `master`, `develop`

**工作流程**:
1. **build-and-test** 任务:
   - 设置 Go 1.24 环境
   - 启动 MySQL 和 PostgreSQL 测试数据库
   - 安装依赖和开发工具
   - 运行 `make all`（构建、测试、代码检查）
   - 运行 MySQL 特定测试
   - 上传测试结果

2. **quality-checks** 任务（仅 PR）:
   - 代码格式检查 (`make fmt-check`)
   - 静态分析 (`go vet`)
   - 代码质量检查 (`golangci-lint`)

3. **security-scan** 任务（仅 master 分支）:
   - 安全扫描 (`gosec`)

4. **release-build** 任务（仅 master 分支推送）:
   - 构建多平台二进制文件
   - 创建发布包
   - 上传构建产物

### 2. `release-build.yml` - 发布构建流水线
**触发条件**: 推送版本标签 (`v*`)

**工作流程**:
1. 运行完整验证 (`make all` + MySQL 测试)
2. 从标签提取版本号
3. 构建多平台发布二进制文件
4. 创建校验和文件
5. 创建 GitHub Release

## 本地开发验证

在提交代码前，建议在本地运行：

```bash
# 运行完整的验证
make all

# 或分步运行
make build    # 构建项目
make test     # 运行测试
make lint     # 代码质量检查
make vet      # 静态分析
make fmt-check # 代码格式检查
```

## 测试数据库

CI 环境中会自动启动：
- **MySQL**: 5.7 版本，数据库 `testdb`，用户 `root`，密码 `rootkit`
- **PostgreSQL**: 17.2 版本，数据库 `testdb`，用户 `postgres`，密码 `rootkit`

## 代码质量要求

### 必须通过的检查
1. **构建**: `go build ./...` 必须成功
2. **测试**: 所有测试必须通过
3. **代码格式**: `make fmt-check` 必须通过
4. **静态分析**: `go vet` 必须通过

### 建议通过的检查
1. **代码质量**: `golangci-lint` 检查
2. **安全扫描**: `gosec` 检查

## 分支策略

### 功能分支 (`feature/*`)
- 从 `develop` 分支创建
- 完成功能后创建 PR 到 `develop`
- 必须通过所有 CI 检查

### 修复分支 (`bugfix/*`, `hotfix/*`)
- 从 `master` 或 `develop` 分支创建
- 完成修复后创建 PR 到对应分支
- 必须通过所有 CI 检查

### 发布流程
1. `develop` 分支准备就绪后，创建 PR 到 `master`
2. 合并到 `master` 后，打上版本标签 (`v1.0.0`)
3. 标签推送会自动触发发布构建

## 故障排除

### 常见问题

#### 1. 本地测试通过但 CI 失败
- 检查测试数据库配置
- 确保没有硬编码的本地路径
- 验证环境变量设置

#### 2. 代码格式检查失败
```bash
# 运行格式化
make fmt

# 检查格式
make fmt-check
```

#### 3. 静态分析失败
```bash
# 运行 vet 检查
go vet ./...

# 查看详细错误
go vet -v ./...
```

#### 4. 构建失败
```bash
# 清理缓存
make clean

# 重新安装依赖
go mod tidy
go mod download
```

## 配置说明

### Makefile 目标
- `make all`: 构建 + 测试 + 代码检查
- `make build`: 构建项目
- `make test`: 运行测试
- `make lint`: 代码质量检查
- `make vet`: 静态分析
- `make fmt`: 格式化代码
- `make fmt-check`: 检查代码格式
- `make clean`: 清理构建文件

### 环境变量
- `GO_VERSION`: Go 版本 (默认: 1.24)
- `MYSQL_VERSION`: MySQL 版本 (默认: 5.7)
- `POSTGRES_VERSION`: PostgreSQL 版本 (默认: 17.2-alpine)

## 性能优化

### 缓存配置
- Go 模块缓存已启用
- 构建缓存已配置
- 测试结果缓存

### 并行执行
- 测试并行执行
- 质量检查并行执行
- 安全扫描并行执行

## 监控和报告

### 测试报告
- 测试覆盖率报告
- 测试结果上传
- 构建产物存档

### 通知
- PR 状态更新
- 构建失败通知
- 发布成功通知

---

**注意**: 所有提交到受保护分支的代码必须通过完整的 CI 流程验证。