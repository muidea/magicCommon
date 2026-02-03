# magicCommon 通用监控框架

[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Test Status](https://img.shields.io/badge/tests-passing-brightgreen.svg)](test/)
[![Production Ready](https://img.shields.io/badge/production-ready-success.svg)](IMPLEMENTATION_COMPLIANCE_REPORT.md)

一个高性能、可扩展的通用监控框架，用于收集、管理和导出应用程序指标。支持Prometheus和JSON格式导出，提供完整的生命周期管理和线程安全保证。

## ✨ 特性

- **多种指标类型**: 支持Counter、Gauge、Histogram、Summary四种标准指标类型
- **模块化设计**: 插件式指标提供者系统，易于扩展
- **高性能**: 异步收集、批量处理、采样控制优化性能
- **线程安全**: 所有组件设计为并发安全，通过并发测试验证
- **多种导出格式**: 支持Prometheus和JSON格式导出
- **安全特性**: HTTP Basic认证、TLS支持、安全头保护
- **环境配置**: 预置开发、生产、高负载环境配置
- **生产就绪**: 通过全面测试验证，具备生产部署条件

## 📦 安装

```bash
go get github.com/muidea/magicCommon/monitoring
```

## 🚀 快速开始

### 基本使用

```go
package main

import (
    "log"
    "time"

    "github.com/muidea/magicCommon/monitoring"
    "github.com/muidea/magicCommon/monitoring/core"
)

func main() {
    // 使用生产环境配置
    config := core.ProductionConfig()
    
    // 创建监控管理器
    manager, err := monitoring.NewManager(&config)
    if err != nil {
        log.Fatalf("Failed to create monitoring manager: %v", err)
    }

    // 启动监控
    if err := manager.Start(); err != nil {
        log.Fatalf("Failed to start monitoring: %v", err)
    }
    defer manager.Shutdown()

    // 业务逻辑...
    time.Sleep(5 * time.Minute)
}
```

### 创建自定义指标提供者

```go
package myapp

import (
    "github.com/muidea/magicCommon/monitoring"
    "github.com/muidea/magicCommon/monitoring/types"
)

type MyMetricsProvider struct {
    *types.BaseProvider
    requestCount int64
}

func NewMyMetricsProvider() *MyMetricsProvider {
    return &MyMetricsProvider{
        BaseProvider: types.NewBaseProvider("myapp", "1.0.0", "My application metrics"),
    }
}

func (p *MyMetricsProvider) Metrics() []types.MetricDefinition {
    return []types.MetricDefinition{
        types.NewCounterDefinition(
            "myapp_requests_total",
            "Total number of requests",
            []string{"method", "endpoint"},
            map[string]string{"app": "myapp"},
        ),
        types.NewGaugeDefinition(
            "myapp_active_users",
            "Number of active users",
            []string{"region"},
            nil,
        ),
    }
}

func (p *MyMetricsProvider) Collect() ([]types.Metric, *types.Error) {
    p.requestCount++
    
    return []types.Metric{
        types.NewCounter(
            "myapp_requests_total",
            float64(p.requestCount),
            map[string]string{
                "method": "GET",
                "endpoint": "/api/users",
            },
        ),
        types.NewGauge(
            "myapp_active_users",
            42.0,
            map[string]string{"region": "us-east-1"},
        ),
    }, nil
}

// 注册提供者
func init() {
    monitoring.RegisterGlobalProvider(
        "myapp",
        func() types.MetricProvider { return NewMyMetricsProvider() },
        true,  // 自动初始化
        100,   // 优先级
    )
}
```

## ⚙️ 配置

### 环境配置

框架提供三种预置环境配置：

```go
// 开发环境配置（低采样率，禁用导出，便于调试）
devConfig := core.DevelopmentConfig()

// 生产环境配置（中等采样率，启用安全导出）
prodConfig := core.ProductionConfig()

// 高负载环境配置（低采样率，优化性能）
highLoadConfig := core.HighLoadConfig()
```

### 自定义配置

```go
config := core.DefaultMonitoringConfig()
config.Namespace = "myapp"
config.SamplingRate = 0.5  // 50%采样率
config.ExportConfig.Enabled = true
config.ExportConfig.Port = 9090
config.ExportConfig.EnablePrometheus = true
config.ExportConfig.EnableAuth = true
config.ExportConfig.AuthToken = "your-secret-token"
```

### 配置验证

```go
if err := config.Validate(); err != nil {
    log.Fatalf("Invalid configuration: %v", err)
}
```

## 📊 指标导出

### Prometheus格式

默认情况下，指标通过HTTP服务器在`/metrics`端点以Prometheus格式提供：

```bash
# 访问指标端点
curl http://localhost:9090/metrics

# 带认证访问
curl -H "Authorization: Bearer your-secret-token" http://localhost:9090/metrics
```

### JSON格式

也可以通过`/api/metrics`端点获取JSON格式的指标：

```bash
curl http://localhost:9090/api/metrics
```

### 健康检查

框架提供健康检查端点：

```bash
curl http://localhost:9090/health
```

## 🔧 API参考

### 监控管理器

```go
// 创建管理器
manager, err := monitoring.NewManager(config)

// 启动监控
err := manager.Start()

// 停止监控
err := manager.Shutdown()

// 注册提供者
err := manager.RegisterProvider(name, factory, autoInit, priority)

// 收集指标
metrics, err := manager.CollectMetrics()

// 导出指标
prometheusData, err := manager.ExportMetrics("prometheus")
jsonData, err := manager.ExportMetrics("json")
```

### 全局管理器

```go
// 初始化全局管理器
monitoring.InitializeGlobalManager()
defer monitoring.ShutdownGlobalManager()

// 注册全局提供者
monitoring.RegisterGlobalProvider(name, factory, autoInit, priority)

// 获取全局管理器
manager := monitoring.GetGlobalManager()
```

### 指标类型

```go
// 创建指标
metric := types.NewMetric(name, metricType, value, labels)

// 创建指标定义
counterDef := types.NewCounterDefinition(name, help, labelNames, constLabels)
gaugeDef := types.NewGaugeDefinition(name, help, labelNames, constLabels)
histogramDef := types.NewHistogramDefinition(name, help, labelNames, buckets, constLabels)
summaryDef := types.NewSummaryDefinition(name, help, labelNames, objectives, maxAge, constLabels)
```

## 🧪 测试

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定测试
go test -v ./test/simple_test.go

# 运行性能基准测试
go test -bench=. ./test/benchmark_test.go

# 运行并发测试
go test -v ./test/concurrency_test.go
```

### 测试覆盖率

```bash
go test -cover ./...
```

## 🛡️ 安全

### 认证

启用HTTP Basic认证：

```go
config.ExportConfig.EnableAuth = true
config.ExportConfig.AuthToken = "your-secret-token"
```

### TLS/HTTPS

启用TLS支持：

```go
config.ExportConfig.EnableTLS = true
config.ExportConfig.TLSCertPath = "/path/to/cert.pem"
config.ExportConfig.TLSKeyPath = "/path/to/key.pem"
```

### 主机白名单

限制访问主机：

```go
config.ExportConfig.AllowedHosts = []string{"127.0.0.1", "localhost"}
```

## 📈 性能优化

### 采样率控制

根据环境调整采样率：

- 开发环境：10%采样率
- 生产环境：50%采样率  
- 高负载环境：10%采样率

### 异步收集

启用异步收集减少业务逻辑阻塞：

```go
config.AsyncCollection = true
config.CollectionInterval = 30 * time.Second
```

### 批量处理

调整批量大小优化性能：

```go
config.BatchSize = 100      // 每批处理100个指标
config.BufferSize = 1000    // 缓冲区大小1000
```

## 🔍 故障排除

### 常见问题

1. **指标未收集**
   - 检查监控是否启用：`config.Enabled = true`
   - 验证采样率配置：`config.SamplingRate > 0`
   - 检查提供者注册状态

2. **内存使用过高**
   - 调整缓冲区大小：`config.BufferSize`
   - 缩短保留时间：`config.RetentionPeriod`
   - 增加采样率：`config.SamplingRate`

3. **导出失败**
   - 检查端口占用：`config.ExportConfig.Port`
   - 验证认证配置：`config.ExportConfig.EnableAuth`
   - 检查网络连接

### 调试工具

1. **健康检查端点**: `/health`
2. **信息端点**: `/`
3. **详细日志**: 启用调试模式
4. **性能分析**: 集成pprof

## 📚 文档

- [设计文档](MONITORING_FRAMEWORK_DESIGN.md) - 详细架构设计和API说明
- [验收要求](ACCEPTANCE_REQUIREMENTS.md) - 功能验收标准和测试要求
- [实现符合性报告](IMPLEMENTATION_COMPLIANCE_REPORT.md) - 实现验证和测试结果
- [测试示例](test/) - 完整的测试用例和示例代码

## 🤝 贡献

欢迎贡献代码、文档和问题报告。请遵循以下步骤：

1. Fork仓库
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建Pull Request

### 开发指南

1. 遵循Go代码规范
2. 添加单元测试
3. 更新相关文档
4. 运行所有测试确保通过

## 📄 许可证

本项目采用MIT许可证。详见[LICENSE](LICENSE)文件。

## 📞 支持

- 问题报告：[GitHub Issues](https://github.com/muidea/magicCommon/issues)
- 文档：[设计文档](MONITORING_FRAMEWORK_DESIGN.md)
- 示例：[测试代码](test/)

---

*最后更新: 2026-02-02*
*版本: 1.0.0*
*状态: 生产就绪*