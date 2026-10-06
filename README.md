# magicCommon

[![CI Pipeline](https://github.com/muidea/magicCommon/actions/workflows/ci.yml/badge.svg)](https://github.com/muidea/magicCommon/actions/workflows/ci.yml)
[![Release Build](https://github.com/muidea/magicCommon/actions/workflows/release-build.yml/badge.svg)](https://github.com/muidea/magicCommon/actions/workflows/release-build.yml)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://golang.org)
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

### Framework 生命周期

`framework/application` 负责进程级 runtime 容器，包括默认
`event.Hub`、`task.BackgroundRoutine`、配置管理器和注入的
`service.Service`。

生产入口推荐统一执行与检查式停机：

```go
err := application.Execute(ctx, service.DefaultService())
```

需要显式配置目录、服务名、队列大小或外部 runtime 组件时，使用
`StartupWithOptions` 或 `NewApplication`：

```go
opts := application.Options{
    ConfigDir:           "./config",
    ServiceName:         "example-service",
    EventHubQueueSize:   1024,
    BackgroundQueueSize: 1024,
}

err := application.StartupWithOptions(ctx, service.DefaultService(), opts)
```

如果注入外部 `EventHub` 或 `BackgroundRoutine`，默认由调用方拥有；
只有在 `Options.Ownership` 中显式声明后，Application 才会在
`Shutdown` 或启动失败清理时终止对应组件。

Application 内部维护 `new`、`starting`、`running`、`failed`、
`stopping`、`shutdown` 状态：

- `Run` 必须在成功 `Startup` 后调用。
- 重复 `Startup` 会返回错误，除非前一次生命周期已经 `Shutdown`。
- `Shutdown` 幂等，成功后保留已关闭的 runtime，不创建新队列/Hub；下一次 `Startup` 才重建默认组件。需要感知失败的调用方使用 `ShutdownChecked(ctx)`。
- 已由 Application 关闭的外部注入对象不能直接复用；重启须通过 `StartupWithOptions` 提供新对象或创建新的 Application。调用方持有 ownership 的外部对象不会被 Application 关闭。
- 启动失败会进入 `failed` 状态并执行 best-effort cleanup；调用方需要
  `Shutdown` 后再重试启动。

`Execute` 封装默认 Application 的 Startup/Run 和最终检查式清理。正常退出、启动/运行失败以及 panic 都先进入清理；每次停机使用独立的 30 秒预算，失败等待 1 秒后重试，不继承已取消的运行 context。完成后返回原执行错误，panic 在清理后继续传播。它不会强制终止未配合取消的回调；这样的回调仍可能阻止进程退出。需要自行控制步骤的调用方仍可使用 Startup/Run/ShutdownChecked。

默认 Service 支持分阶段停机：全部已进入 Setup 的插件先执行可选 `BeginShutdown(context.Context)` 关闭入口/请求取消，再执行 `Quiesce(context.Context) *def.Error` 等待在途操作；Application 随后确认后台队列、定时器和已接受的事件排空，才执行最终 `Teardown` 及 Hub 关闭。失败或 panic 返回错误并保留尚未释放的依赖，状态为 stopping，可再次尝试停机，不能重新 Run/Startup。最终清理按逆序遇错即停，重试不会重复执行已成功的阶段；已完成的释放不承诺回滚。自定义 Service 可实现 `service.Quiescer` 和 `service.CheckedShutdown` 接入屏障和最终清理回执；未实现 Quiescer 的 Service 必须在自己的 Shutdown 中先完成入口关闭与资源排空。

插件 Setup 失败不再由单个 PluginMgr 提前回滚：已进入的插件（包括部分失败的插件）保留给进程级统一停机流程，未进入 Setup 的插件不参与清理。Application 在启动失败时仍自动尝试 `ShutdownChecked`；直接使用 DefaultService/PluginMgr 的调用方须显式执行检查式清理。插件需保证部分 Setup 后也可安全、幂等地清理。最终 Teardown 的错误及 panic、LifecycleService adapter 的清理错误均向上返回。

`BackgroundRoutine.AsyncTaskContext` 的 context 约束入队等待；已接受的任务仍需由 owner 在执行时处理取消及释放回执。定时器首次和后续 tick 都使用此入口；注册拒绝已取消上下文和已关闭调度器，关闭会唤醒阻塞提交者并等待 timer goroutine 退出。注册成功不等于未来每次 tick 都已完成。

同步 EventHub Send 保留仅在本 Hub、当前调用链仍活跃时有效的祖先通道，支持 A → B → A；独立调用链形成循环依赖时明确拒绝。排队中尚未执行的 Send 可取消，取消成功保证不再执行 handler；已经执行的回调必须等真实返回。不要把同步派发上下文用于并发调用或延迟重入。Post 总是排队，不能继承祖先重入权限。`event.DrainingHub.TerminateChecked` 拒绝新根调用，保留在途同步依赖，超时不清空订阅，支持继续排空和重试；旧的 void Terminate 入口仅记录未完成错误。

`framework/service` 还提供 foreground lifecycle adapter：

```go
type localService struct{}

func (s *localService) Startup(ctx context.Context) error { return nil }
func (s *localService) Run(ctx context.Context) error     { return nil }
func (s *localService) Shutdown(ctx context.Context) error { return nil }

svc := service.AdaptLifecycle("local", &localService{})
err := application.Startup(ctx, svc)
```

Plugin 注册建议优先使用可观测 API：

```go
err := initiator.RegisterE(plugin)
module.MustRegister(plugin)
```

`Register` 仍保留原签名并记录注册错误；`RegisterE` 返回重复 ID、nil、
非指针、缺失 `ID` / `Run` 或方法签名不匹配等错误。新插件可实现
`framework/plugin/common` 中的显式接口，旧插件的反射兼容路径仍可用。

### 环境要求
- Go 1.26+
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

1. 完成升级说明及验证，将代码提交到 `master`。
2. 创建并推送不可覆写的版本标签（例如 `v1.5.16`），供 Go module 按版本引用。
3. 标签触发 `release-build.yml` 的构建、质量检查及 PostgreSQL/MySQL 测试。
4. 当前工作流不自动创建 GitHub Release 条目，也不生成多平台二进制发布包。

当前版本为 `v1.5.21`，补充框架与业务模型边界维护规则；框架保持通用能力，平台业务流程记录由应用项目负责。升级时使用 `go get github.com/muidea/magicCommon@v1.5.21`，再同步调用项目的 vendor。

此前 `v1.5.20`：新增 `foundation/net.WriteFileAtomic`，仅在完整复制、同步并关闭文件后原子提交；HTTP 与 multipart 上传复用该实现，并正确返回保留策略锁与临时文件清理错误。`v1.5.20` 同时修正 MySQL 连接池压力测试：各并发操作独立持有 DAO 查询游标，共享底层连接池，并校验全部插入完成。此前变更见 [v1.5.18 发布说明](release-note-v1.5.18.md)；从旧版本升级时仍须核对 [v1.5.16 接口与生命周期升级要求](release-note-v1.5.16.md)。

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
- [framework-lifecycle-improvement-plan.md](./framework-lifecycle-improvement-plan.md): framework lifecycle、plugin 注册安全、Application options 与 lifecycle adapter 收口记录
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

## 按窗口观测调用成本

`foundation/profiling` 提供默认关闭的有界进程窗口统计。启动时设置 `MAGIC_PROFILE_WINDOW=60s` 可记录公共 HTTP helper 的调用次数、业务/传输错误、累计时间及耗时桶；不记录 URL 主机、动态标识或参数。标签有上限，并行调用累计时间不是 CPU 时间。配置与采集说明见工作区 [QPS 分析指南](../magicRunner/docs/guide-qps-analysis.md)。

`foundation/retention` 提供与业务无关的日志保留限制、版本化配置文件、原子保存和后台读取调度；不包含应用权限、数据查询或删除逻辑。
