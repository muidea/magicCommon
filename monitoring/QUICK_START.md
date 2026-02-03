# 快速开始指南

本指南将帮助你在5分钟内启动并使用magicCommon监控框架。

## 🎯 目标

通过本指南，你将学会：
1. 安装和导入监控框架
2. 创建和配置监控管理器
3. 创建自定义指标提供者
4. 查看和导出监控指标
5. 在生产环境中部署

## 📦 步骤1：安装

### 使用go get安装

```bash
go get github.com/muidea/magicCommon/monitoring
```

### 在Go模块中导入

```go
import (
    "github.com/muidea/magicCommon/monitoring"
    "github.com/muidea/magicCommon/monitoring/core"
    "github.com/muidea/magicCommon/monitoring/types"
)
```

## 🚀 步骤2：基本使用

### 最简单的示例

创建一个最简单的监控应用：

```go
package main

import (
    "log"
    "time"

    "github.com/muidea/magicCommon/monitoring"
    "github.com/muidea/magicCommon/monitoring/core"
)

func main() {
    // 1. 使用默认配置
    config := core.DefaultMonitoringConfig()
    
    // 2. 创建监控管理器
    manager, err := monitoring.NewManager(&config)
    if err != nil {
        log.Fatalf("Failed to create manager: %v", err)
    }

    // 3. 启动监控
    if err := manager.Start(); err != nil {
        log.Fatalf("Failed to start monitoring: %v", err)
    }
    defer manager.Shutdown()

    log.Println("Monitoring started successfully!")
    
    // 4. 保持程序运行以查看指标
    time.Sleep(10 * time.Minute)
}
```

运行此程序后，访问 http://localhost:9090/metrics 查看Prometheus格式的指标。

## 📊 步骤3：创建第一个指标提供者

### 创建简单的计数器提供者

```go
package main

import (
    "log"
    "time"

    "github.com/muidea/magicCommon/monitoring"
    "github.com/muidea/magicCommon/monitoring/core"
    "github.com/muidea/magicCommon/monitoring/types"
)

// SimpleCounterProvider 是一个简单的计数器提供者
type SimpleCounterProvider struct {
    *types.BaseProvider
    count int64
}

func NewSimpleCounterProvider() *SimpleCounterProvider {
    return &SimpleCounterProvider{
        BaseProvider: types.NewBaseProvider("simple", "1.0.0", "Simple counter metrics"),
        count: 0,
    }
}

func (p *SimpleCounterProvider) Metrics() []types.MetricDefinition {
    return []types.MetricDefinition{
        types.NewCounterDefinition(
            "simple_counter_total",
            "Total count of simple operations",
            []string{"operation"},
            map[string]string{"type": "demo"},
        ),
    }
}

func (p *SimpleCounterProvider) Collect() ([]types.Metric, *types.Error) {
    p.count++
    
    return []types.Metric{
        types.NewCounter(
            "simple_counter_total",
            float64(p.count),
            map[string]string{"operation": "increment"},
        ),
    }, nil
}

func main() {
    // 使用开发环境配置
    config := core.DevelopmentConfig()
    
    // 创建管理器
    manager, err := monitoring.NewManager(&config)
    if err != nil {
        log.Fatalf("Failed to create manager: %v", err)
    }

    // 注册自定义提供者
    err = manager.RegisterProvider(
        "simple",
        func() types.MetricProvider { return NewSimpleCounterProvider() },
        true,  // 自动初始化
        100,   // 优先级
    )
    if err != nil {
        log.Fatalf("Failed to register provider: %v", err)
    }

    // 启动监控
    if err := manager.Start(); err != nil {
        log.Fatalf("Failed to start monitoring: %v", err)
    }
    defer manager.Shutdown()

    log.Println("Simple counter provider started!")
    
    // 模拟业务逻辑
    for i := 0; i < 60; i++ {
        time.Sleep(1 * time.Second)
        log.Printf("Counter value: %d", i+1)
    }
}
```

## ⚙️ 步骤4：配置详解

### 环境配置选择

根据你的环境选择合适的配置：

```go
// 开发环境 - 低采样率，便于调试
config := core.DevelopmentConfig()

// 生产环境 - 中等采样率，启用安全特性
config := core.ProductionConfig()

// 高负载环境 - 低采样率，优化性能
config := core.HighLoadConfig()
```

### 自定义配置示例

```go
config := core.DefaultMonitoringConfig()

// 基本配置
config.Namespace = "myapp"           // 指标命名空间
config.SamplingRate = 0.5            // 50%采样率
config.AsyncCollection = true        // 启用异步收集
config.CollectionInterval = 30 * time.Second

// 导出配置
config.ExportConfig.Enabled = true
config.ExportConfig.Port = 9090
config.ExportConfig.Path = "/metrics"
config.ExportConfig.EnablePrometheus = true
config.ExportConfig.EnableJSON = true

// 安全配置
config.ExportConfig.EnableAuth = true
config.ExportConfig.AuthToken = "my-secret-token"

// 性能优化
config.BatchSize = 100
config.BufferSize = 1000
config.MaxConcurrentTasks = 10
```

### 配置验证

```go
if err := config.Validate(); err != nil {
    log.Fatalf("Configuration validation failed: %v", err)
}
```

## 🔧 步骤5：高级特性

### 使用全局管理器

```go
package main

import (
    "log"
    "time"

    "github.com/muidea/magicCommon/monitoring"
)

func main() {
    // 初始化全局管理器
    monitoring.InitializeGlobalManager()
    defer monitoring.ShutdownGlobalManager()

    // 现在可以在任何地方注册提供者
    // 提供者会在管理器启动时自动初始化
    
    // 获取全局管理器进行操作
    manager := monitoring.GetGlobalManager()
    if err := manager.Start(); err != nil {
        log.Fatalf("Failed to start global manager: %v", err)
    }

    time.Sleep(5 * time.Minute)
}
```

### 创建多种指标类型

```go
package metrics

import (
    "github.com/muidea/magicCommon/monitoring/types"
)

type MultiMetricProvider struct {
    *types.BaseProvider
    counter   int64
    gaugeVal  float64
    histogram []float64
}

func NewMultiMetricProvider() *MultiMetricProvider {
    return &MultiMetricProvider{
        BaseProvider: types.NewBaseProvider("multi", "1.0.0", "Multiple metric types"),
        counter:      0,
        gaugeVal:     100.0,
        histogram:    []float64{0.1, 0.5, 0.9, 1.2, 2.5},
    }
}

func (p *MultiMetricProvider) Metrics() []types.MetricDefinition {
    return []types.MetricDefinition{
        // 计数器
        types.NewCounterDefinition(
            "multi_requests_total",
            "Total requests processed",
            []string{"method", "status"},
            nil,
        ),
        
        // 仪表盘
        types.NewGaugeDefinition(
            "multi_memory_usage_bytes",
            "Current memory usage in bytes",
            []string{"type"},
            nil,
        ),
        
        // 直方图
        types.NewHistogramDefinition(
            "multi_request_duration_seconds",
            "Request duration in seconds",
            []string{"endpoint"},
            []float64{0.1, 0.5, 1.0, 2.0, 5.0}, // 桶边界
            nil,
        ),
        
        // 摘要
        types.NewSummaryDefinition(
            "multi_response_size_bytes",
            "Response size in bytes",
            []string{"content_type"},
            map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}, // 分位数目标
            10 * time.Minute, // 最大年龄
            nil,
        ),
    }
}

func (p *MultiMetricProvider) Collect() ([]types.Metric, *types.Error) {
    p.counter++
    p.gaugeVal += 0.5
    
    return []types.Metric{
        // 计数器指标
        types.NewCounter(
            "multi_requests_total",
            float64(p.counter),
            map[string]string{
                "method": "POST",
                "status": "200",
            },
        ),
        
        // 仪表盘指标
        types.NewGauge(
            "multi_memory_usage_bytes",
            p.gaugeVal,
            map[string]string{"type": "heap"},
        ),
        
        // 直方图指标（多个观测值）
        types.NewHistogram(
            "multi_request_duration_seconds",
            p.histogram,
            map[string]string{"endpoint": "/api/users"},
        ),
        
        // 摘要指标
        types.NewSummary(
            "multi_response_size_bytes",
            1024.0, // 观测值
            map[string]string{"content_type": "application/json"},
        ),
    }, nil
}
```

## 📈 步骤6：查看和导出指标

### 通过HTTP访问指标

启动监控后，可以通过以下端点访问指标：

```bash
# Prometheus格式（默认）
curl http://localhost:9090/metrics

# JSON格式
curl http://localhost:9090/api/metrics

# 健康检查
curl http://localhost:9090/health

# 系统信息
curl http://localhost:9090/
```

### 带认证访问

如果启用了认证：

```bash
# 使用Bearer Token
curl -H "Authorization: Bearer your-secret-token" http://localhost:9090/metrics

# 或使用Basic Auth
curl -u username:password http://localhost:9090/metrics
```

### 通过代码导出

```go
// 导出为Prometheus格式
prometheusData, err := manager.ExportMetrics("prometheus")
if err == nil {
    fmt.Println("Prometheus metrics:", prometheusData)
}

// 导出为JSON格式
jsonData, err := manager.ExportMetrics("json")
if err == nil {
    fmt.Println("JSON metrics:", jsonData)
}
```

## 🧪 步骤7：测试你的实现

### 运行示例测试

```bash
# 进入monitoring目录
cd monitoring

# 运行简单测试
go test -v ./test/simple_test.go

# 运行所有测试
go test ./...

# 检查测试覆盖率
go test -cover ./...
```

### 创建单元测试

```go
package myapp_test

import (
    "testing"
    "time"

    "github.com/muidea/magicCommon/monitoring"
    "github.com/muidea/magicCommon/monitoring/core"
    "github.com/stretchr/testify/assert"
)

func TestMonitoringBasic(t *testing.T) {
    // 使用测试配置
    config := core.DevelopmentConfig()
    config.ExportConfig.Enabled = false // 测试时禁用导出
    
    manager, err := monitoring.NewManager(&config)
    assert.NoError(t, err)
    assert.NotNil(t, manager)
    
    // 测试启动和停止
    err = manager.Start()
    assert.NoError(t, err)
    
    time.Sleep(100 * time.Millisecond)
    
    err = manager.Shutdown()
    assert.NoError(t, err)
}

func TestCustomProvider(t *testing.T) {
    config := core.DevelopmentConfig()
    config.ExportConfig.Enabled = false
    
    manager, err := monitoring.NewManager(&config)
    assert.NoError(t, err)
    
    // 测试提供者注册和收集
    // ...
}
```

## 🚀 步骤8：生产部署

### 生产环境配置建议

```go
func getProductionConfig() core.MonitoringConfig {
    config := core.ProductionConfig()
    
    // 根据业务需求调整
    config.Namespace = "my-production-app"
    config.SamplingRate = 0.3 // 30%采样率，平衡性能和数据完整性
    
    // 安全配置
    config.ExportConfig.EnableAuth = true
    config.ExportConfig.AuthToken = getSecureToken() // 从安全存储获取
    
    // 性能优化
    config.BatchSize = 200
    config.BufferSize = 5000
    config.MaxConcurrentTasks = 20
    
    // 监控框架自身
    config.ProviderConfigs = map[string]interface{}{
        "monitoring": map[string]interface{}{
            "enable_self_monitoring": true,
            "collection_interval": "30s",
        },
    }
    
    return config
}
```

### 部署检查清单

- [ ] 验证配置正确性
- [ ] 测试认证和授权
- [ ] 验证指标导出功能
- [ ] 监控框架自身运行状态
- [ ] 设置适当的告警规则
- [ ] 配置日志记录和审计

## 🆘 故障排除

### 常见问题

1. **端口冲突**
   ```
   Error: listen tcp :9090: bind: address already in use
   ```
   解决方案：更改端口号或停止占用端口的进程

2. **认证失败**
   ```
   Error: 401 Unauthorized
   ```
   解决方案：检查认证令牌配置

3. **内存使用过高**
   解决方案：调整缓冲区大小和采样率

### 获取帮助

- 查看[设计文档](MONITORING_FRAMEWORK_DESIGN.md)了解详细架构
- 查看[实现符合性报告](IMPLEMENTATION_COMPLIANCE_REPORT.md)了解测试结果
- 查看[测试示例](test/)学习更多用法
- 提交[GitHub Issues](https://github.com/muidea/magicCommon/issues)报告问题

## 🎉 恭喜！

你已经成功学会了：
- ✅ 安装和导入监控框架
- ✅ 创建和配置监控管理器  
- ✅ 创建自定义指标提供者
- ✅ 查看和导出监控指标
- ✅ 测试和部署监控系统

现在你可以开始监控你的应用程序了！继续探索框架的高级特性，或查看其他文档了解更多细节。

---

*快速开始指南版本: 1.0.0*
*最后更新: 2026-02-02*
*下一步建议: 查看[最佳实践指南](BEST_PRACTICES.md)学习生产环境部署技巧*