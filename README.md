# magicCommon

[![CI Pipeline](https://github.com/muidea/magicCommon/actions/workflows/ci.yml/badge.svg)](https://github.com/muidea/magicCommon/actions/workflows/ci.yml)
[![Release Build](https://github.com/muidea/magicCommon/actions/workflows/release-build.yml/badge.svg)](https://github.com/muidea/magicCommon/actions/workflows/release-build.yml)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

一个 Go 语言库，提供通用的工具、基础框架和应用构建模块。

## 功能特性

### 基础工具 (Foundation)
- **缓存系统**: 内存缓存、分布式缓存支持
- **对象池**: 泛型资源池、预创建与复用控制
- **数据访问层**: MySQL、PostgreSQL DAO 实现
- **日志系统**: 结构化日志、多输出支持
- **网络工具**: HTTP 客户端、服务器工具
- **路径工具**: 路径遍历、目录复制与文件监控
- **同步信号**: 进程内轻量信号协调
- **系统工具**: 文件操作、进程管理
- **工具函数**: 字符串、时间、加密等工具

### 框架组件 (Framework)
- **应用框架**: 应用生命周期管理
- **配置管理**: 多格式配置、热重载
- **插件系统**: 模块化插件架构
- **服务框架**: 微服务基础组件

### 其他模块
- **事件系统**: 发布-订阅模式
- **执行器**: 并发任务执行与等待控制
- **监控系统**: 指标收集、监控集成
- **会话管理**: 用户会话管理
- **任务调度**: 定时任务、异步任务

## 快速开始

### 安装

```bash
go get github.com/muidea/magicCommon
```

### 使用示例

```go
package main

import (
    "fmt"
    
    cd "github.com/muidea/magicCommon/def"
    "github.com/muidea/magicCommon/foundation/dao"
    "github.com/muidea/magicCommon/foundation/log"
)

func main() {
    // 初始化日志
    log.Infof("Starting application...")
    
    // 创建数据库连接
    db, err := dao.Fetch("user", "password", "localhost:3306", "testdb")
    if err != nil {
        log.Errorf("Failed to connect to database: %v", err)
        return
    }
    defer db.Release()
    
    // 执行查询
    err = db.Query("SELECT 1")
    if err != nil {
        log.Errorf("Query failed: %v", err)
        return
    }
    
    log.Infof("Application started successfully")
}
```

## 开发指南

### 环境要求
- Go 1.24+
- MySQL 5.7+ (用于测试)
- PostgreSQL 17.2+ (用于测试)

### 本地开发

```bash
# 克隆项目
git clone https://github.com/muidea/magicCommon.git
cd magicCommon

# 安装依赖
go mod tidy
go mod download

# 运行完整验证
make all

# 运行测试
make test

# 代码质量检查
make lint
```

### 测试数据库

项目测试需要 MySQL 和 PostgreSQL 数据库。可以使用 Docker 快速启动：

```bash
# 启动 MySQL
docker run -d --name mysql-test \
  -e MYSQL_DATABASE=testdb \
  -e MYSQL_ROOT_PASSWORD=rootkit \
  -p 3306:3306 \
  mysql:5.7

# 启动 PostgreSQL
docker run -d --name postgres-test \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=rootkit \
  -e POSTGRES_DB=testdb \
  -p 5432:5432 \
  postgres:17.2-alpine
```

## CI/CD 流程

项目使用 GitHub Actions 进行持续集成和持续部署：

### 提交代码前
```bash
# 运行完整验证
make all

# 或分步运行
make build    # 构建项目
make test     # 运行测试  
make lint     # 代码质量检查
make vet      # 静态分析
make fmt-check # 代码格式检查
```

### CI 流程
1. **代码推送** 或 **PR 创建** 触发 CI
2. 运行 `make all`（构建、测试、代码检查）
3. 运行 MySQL 特定测试
4. 代码质量检查（格式、静态分析）
5. 安全扫描（仅 master 分支）
6. 构建发布二进制文件（仅 master 分支）

### 发布流程
1. 代码合并到 `master` 分支
2. 创建版本标签 (`v1.0.0`)
3. 自动触发发布构建
4. 生成多平台二进制文件
5. 创建 GitHub Release

## 项目结构

```
magicCommon/
├── def/              # 通用定义和错误类型
├── foundation/       # 基础工具
│   ├── cache/       # 缓存系统
│   ├── dao/         # 数据访问层
│   ├── log/         # 日志系统
│   ├── net/         # 网络工具
│   ├── os/          # 系统工具
│   ├── path/        # 路径工具
│   ├── pool/        # 连接池
│   ├── signal/      # 信号处理
│   ├── system/      # 系统工具
│   └── util/        # 工具函数
├── framework/       # 框架组件
│   ├── application/ # 应用框架
│   ├── configuration/ # 配置管理
│   ├── plugin/      # 插件系统
│   └── service/     # 服务框架
├── event/           # 事件系统
├── execute/         # 并发执行器
├── monitoring/      # 监控系统
├── session/         # 会话管理
├── task/            # 任务调度
└── test/            # 测试工具
```

## 核心基础设施文档

- [technical-note-infra-hardening-2026-03.md](./technical-note-infra-hardening-2026-03.md): 本轮基础设施修复与稳定语义总结
- [release-note-2026-03-lifecycle-cache-monitoring.md](./release-note-2026-03-lifecycle-cache-monitoring.md): 本轮 lifecycle、cache、monitoring 变化摘要
- [event/README.md](./event/README.md): 事件中心、投递、关闭与匹配语义
- [execute/README.md](./execute/README.md): 执行器并发限制与等待语义
- [task/README.md](./task/README.md): 后台任务与超时等待语义
- [foundation/cache/README.md](./foundation/cache/README.md): 内存缓存、过期清理与释放语义
- [foundation/net/README.md](./foundation/net/README.md): HTTP helper、文件上传下载与 DNS client 语义
- [foundation/pool/README.md](./foundation/pool/README.md): 泛型资源池、关闭与预创建语义
- [foundation/path/README.md](./foundation/path/README.md): 路径工具与目录监控语义
- [foundation/signal/README.md](./foundation/signal/README.md): 进程内信号协调与关闭语义
- [framework/configuration/README.md](./framework/configuration/README.md): 配置框架与 watcher 行为
- [monitoring/README.md](./monitoring/README.md): 监控系统入口说明

## 代码规范

### 导入顺序
1. 标准库导入
2. 第三方库导入
3. 项目内部导入

### 错误处理
- 使用 `*cd.Error` 类型处理错误
- 错误变量命名为 `err` 或 `errVal`
- 使用 `cd.NewError()` 创建错误

### 测试规范
- 测试文件命名为 `*_test.go`
- 测试函数以 `Test` 开头
- 使用表驱动测试
- 使用 `t.Run()` 组织子测试

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 提交信息规范
- feat: 新功能
- fix: 修复 bug
- docs: 文档更新
- style: 代码格式调整
- refactor: 代码重构
- test: 测试相关
- chore: 构建过程或辅助工具变动

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 联系方式

- 项目地址: https://github.com/muidea/magicCommon
- 问题反馈: https://github.com/muidea/magicCommon/issues
- 讨论区: https://github.com/muidea/magicCommon/discussions

## 致谢

感谢所有为这个项目做出贡献的开发者！

---

**提示**: 在提交代码前，请确保运行 `make all` 通过所有检查。
