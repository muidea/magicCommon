# API 参考文档

本文档提供magicCommon监控框架的完整API参考。

## 📋 目录

- [包结构](#包结构)
- [监控管理器](#监控管理器)
- [配置管理](#配置管理)
- [指标类型](#指标类型)
- [指标提供者](#指标提供者)
- [指标收集器](#指标收集器)
- [注册表管理](#注册表管理)
- [指标导出器](#指标导出器)
- [错误处理](#错误处理)
- [全局管理器](#全局管理器)

## 📦 包结构

### 主包：`monitoring`
```go
import "github.com/muidea/magicCommon/monitoring"
```
监控框架的主入口点，提供监控管理器。

### 核心包：`monitoring/core`
```go
import "github.com/muidea/magicCommon/monitoring/core"
```
包含核心组件：配置、收集器、注册表、导出器。

### 类型包：`monitoring/types`
```go
import "github.com/muidea/magicCommon/monitoring/types"
```
包含指标类型、定义、提供者接口和错误类型。

## 🎮 监控管理器

### Manager 结构体

```go
type Manager struct {
    // 内部字段，外部不应直接访问
}
```

### 构造函数

#### NewManager
```go
func NewManager(config *core.MonitoringConfig) (*Manager, *types.Error)
```
创建新的监控管理器。

**参数：**
- `config` - 监控配置，如果为nil则使用默认配置

**返回值：**
- `*Manager` - 监控管理器实例
- `*types.Error` - 错误信息，如果创建失败

**示例：**
```go
config := core.DefaultMonitoringConfig()
manager, err := monitoring.NewManager(&config)
if err != nil {
    log.Fatal(err)
}
```

### 方法

#### Start
```go
func (m *Manager) Start() *types.Error
```
启动监控管理器，开始收集和导出指标。

**返回值：**
- `*types.Error` - 错误信息，如果启动失败

#### Shutdown
```go
func (m *Manager) Shutdown() *types.Error
```
停止监控管理器，清理所有资源。

**返回值：**
- `*types.Error` - 错误信息，如果停止失败

#### RegisterProvider
```go
func (m *Manager) RegisterProvider(
    name string,
    factory types.ProviderFactory,
    autoInit bool,
    priority int,
) *types.Error
```
注册指标提供者。

**参数：**
- `name` - 提供者名称（唯一标识）
- `factory` - 提供者工厂函数
- `autoInit` - 是否自动初始化
- `priority` - 初始化优先级（数值越小优先级越高）

**返回值：**
- `*types.Error` - 错误信息，如果注册失败

**示例：**
```go
err := manager.RegisterProvider(
    "myapp",
    func() types.MetricProvider { return NewMyAppProvider() },
    true,  // 自动初始化
    100,   // 优先级
)
```

#### CollectMetrics
```go
func (m *Manager) CollectMetrics() ([]types.Metric, *types.Error)
```
手动触发指标收集。

**返回值：**
- `[]types.Metric` - 收集到的指标列表
- `*types.Error` - 错误信息，如果收集失败

#### ExportMetrics
```go
func (m *Manager) ExportMetrics(format string) (string, *types.Error)
```
导出指标为指定格式。

**参数：**
- `format` - 导出格式，支持 "prometheus" 或 "json"

**返回值：**
- `string` - 格式化后的指标数据
- `*types.Error` - 错误信息，如果导出失败

**示例：**
```go
// 导出为Prometheus格式
prometheusData, err := manager.ExportMetrics("prometheus")
if err == nil {
    fmt.Println(prometheusData)
}

// 导出为JSON格式
jsonData, err := manager.ExportMetrics("json")
if err == nil {
    fmt.Println(jsonData)
}
```

#### GetStats
```go
func (m *Manager) GetStats() ManagerStats
```
获取管理器统计信息。

**返回值：**
- `ManagerStats` - 管理器统计信息

### ManagerStats 结构体

```go
type ManagerStats struct {
    StartTime        int64 `json:"start_time"`         // 启动时间戳（Unix秒）
    UptimeSeconds    int64 `json:"uptime_seconds"`     // 运行时间（秒）
    TotalMetrics     int64 `json:"total_metrics"`      // 总指标数量
    ActiveProviders  int64 `json:"active_providers"`   // 活跃提供者数量
    ExportRequests   int64 `json:"export_requests"`    // 导出请求次数
    CollectionCycles int64 `json:"collection_cycles"`  // 收集周期次数
}
```

## ⚙️ 配置管理

### MonitoringConfig 结构体

```go
type MonitoringConfig struct {
    Enabled            bool          `json:"enabled"`              // 是否启用监控
    Namespace          string        `json:"namespace"`            // 指标命名空间前缀
    SamplingRate       float64       `json:"sampling_rate"`        // 采样率 (0.0-1.0)
    AsyncCollection    bool          `json:"async_collection"`     // 是否异步收集
    CollectionInterval time.Duration `json:"collection_interval"`  // 收集间隔
    RetentionPeriod    time.Duration `json:"retention_period"`     // 指标保留时间
    DetailLevel        DetailLevel   `json:"detail_level"`         // 详细级别
    ExportConfig       ExportConfig  `json:"export_config"`        // 导出配置
    BatchSize          int           `json:"batch_size"`           // 批量大小
    BufferSize         int           `json:"buffer_size"`          // 缓冲区大小
    MaxConcurrentTasks int           `json:"max_concurrent_tasks"` // 最大并发任务数
    Timeout            time.Duration `json:"timeout"`              // 操作超时时间
    ProviderConfigs    map[string]interface{} `json:"provider_configs,omitempty"` // 提供者特定配置
    Environment        string        `json:"environment"`          // 环境标识
}
```

### ExportConfig 结构体

```go
type ExportConfig struct {
    Enabled         bool          `json:"enabled"`           // 是否启用导出
    Port            int           `json:"port"`              // HTTP服务器端口
    Path            string        `json:"path"`              // Prometheus指标路径
    HealthCheckPath string        `json:"health_check_path"` // 健康检查路径
    MetricsPath     string        `json:"metrics_path"`      // JSON指标路径
    InfoPath        string        `json:"info_path"`         // 信息路径
    EnablePrometheus bool         `json:"enable_prometheus"` // 启用Prometheus格式
    EnableJSON      bool          `json:"enable_json"`       // 启用JSON格式
    RefreshInterval time.Duration `json:"refresh_interval"`  // 刷新间隔
    ScrapeTimeout   time.Duration `json:"scrape_timeout"`    // 抓取超时时间
    EnableTLS       bool          `json:"enable_tls"`        // 启用TLS
    TLSCertPath     string        `json:"tls_cert_path"`     // TLS证书路径
    TLSKeyPath      string        `json:"tls_key_path"`      // TLS密钥路径
    EnableAuth      bool          `json:"enable_auth"`       // 启用认证
    AuthToken       string        `json:"auth_token"`        // 认证令牌
    AllowedHosts    []string      `json:"allowed_hosts"`     // 允许的主机列表
}
```

### DetailLevel 类型

```go
type DetailLevel string

const (
    DetailLevelBasic    DetailLevel = "basic"    // 仅收集基本指标
    DetailLevelStandard DetailLevel = "standard" // 收集标准操作指标
    DetailLevelDetailed DetailLevel = "detailed" // 收集详细指标（包括性能分析）
)
```

### 配置函数

#### DefaultMonitoringConfig
```go
func DefaultMonitoringConfig() MonitoringConfig
```
返回默认监控配置。

#### DefaultExportConfig
```go
func DefaultExportConfig() ExportConfig
```
返回默认导出配置。

#### DevelopmentConfig
```go
func DevelopmentConfig() MonitoringConfig
```
返回开发环境配置（低采样率，禁用导出，便于调试）。

#### ProductionConfig
```go
func ProductionConfig() MonitoringConfig
```
返回生产环境配置（中等采样率，启用安全导出）。

#### HighLoadConfig
```go
func HighLoadConfig() MonitoringConfig
```
返回高负载环境配置（低采样率，优化性能）。

#### Validate
```go
func (c *MonitoringConfig) Validate() *types.Error
```
验证配置完整性。

#### MergeConfigs
```go
func MergeConfigs(base, override MonitoringConfig) MonitoringConfig
```
合并两个配置，override中的值会覆盖base中的值。

## 📊 指标类型

### MetricType 类型

```go
type MetricType string

const (
    CounterMetric   MetricType = "counter"   // 计数器：只增不减
    GaugeMetric     MetricType = "gauge"     // 仪表盘：可增可减
    HistogramMetric MetricType = "histogram" // 直方图：采样观测值
    SummaryMetric   MetricType = "summary"   // 摘要：计算分位数
)
```

### Metric 结构体

```go
type Metric struct {
    Name        string            `json:"name"`                  // 指标名称
    Type        MetricType        `json:"type"`                  // 指标类型
    Value       float64           `json:"value"`                 // 指标值
    Labels      map[string]string `json:"labels"`                // 标签键值对
    Timestamp   time.Time         `json:"timestamp"`             // 时间戳
    Description string            `json:"description,omitempty"` // 描述（可选）
}
```

### MetricDefinition 结构体

```go
type MetricDefinition struct {
    Name        string              `json:"name"`                    // 指标名称
    Type        MetricType          `json:"type"`                    // 指标类型
    Help        string              `json:"help"`                    // 帮助文本
    LabelNames  []string            `json:"label_names"`             // 标签名称列表
    Buckets     []float64           `json:"buckets,omitempty"`       // 直方图桶边界
    Objectives  map[float64]float64 `json:"objectives,omitempty"`    // 摘要分位数目标
    MaxAge      time.Duration       `json:"max_age,omitempty"`       // 最大年龄
    ConstLabels map[string]string   `json:"const_labels,omitempty"`  // 常量标签
}
```

### 指标创建函数

#### NewMetric
```go
func NewMetric(name string, metricType MetricType, value float64, labels map[string]string) Metric
```
创建新的指标实例。

#### NewCounter
```go
func NewCounter(name string, value float64, labels map[string]string) Metric
```
创建计数器指标。

#### NewGauge
```go
func NewGauge(name string, value float64, labels map[string]string) Metric
```
创建仪表盘指标。

#### NewHistogram
```go
func NewHistogram(name string, values []float64, labels map[string]string) Metric
```
创建直方图指标（多个观测值）。

#### NewSummary
```go
func NewSummary(name string, value float64, labels map[string]string) Metric
```
创建摘要指标。

### 指标定义创建函数

#### NewCounterDefinition
```go
func NewCounterDefinition(
    name string,
    help string,
    labelNames []string,
    constLabels map[string]string,
) MetricDefinition
```
创建计数器定义。

#### NewGaugeDefinition
```go
func NewGaugeDefinition(
    name string,
    help string,
    labelNames []string,
    constLabels map[string]string,
) MetricDefinition
```
创建仪表盘定义。

#### NewHistogramDefinition
```go
func NewHistogramDefinition(
    name string,
    help string,
    labelNames []string,
    buckets []float64,
    constLabels map[string]string,
) MetricDefinition
```
创建直方图定义。

#### NewSummaryDefinition
```go
func NewSummaryDefinition(
    name string,
    help string,
    labelNames []string,
    objectives map[float64]float64,
    maxAge time.Duration,
    constLabels map[string]string,
) MetricDefinition
```
创建摘要定义。

### 指标定义方法

#### Validate
```go
func (d *MetricDefinition) Validate() *types.Error
```
验证指标定义的有效性。

#### GetFullName
```go
func (d *MetricDefinition) GetFullName(namespace string) string
```
获取带命名空间的完整指标名称。

## 🔌 指标提供者

### MetricProvider 接口

```go
type MetricProvider interface {
    Name() string                           // 提供者名称
    Metrics() []MetricDefinition            // 提供的指标定义
    Init(collector interface{}) *types.Error // 初始化提供者
    Collect() ([]Metric, *types.Error)      // 收集指标
    Shutdown() *types.Error                 // 关闭提供者
    GetMetadata() ProviderMetadata          // 获取元数据
}
```

### ProviderFactory 类型

```go
type ProviderFactory func() MetricProvider
```
提供者工厂函数类型。

### ProviderMetadata 结构体

```go
type ProviderMetadata struct {
    Name         string    `json:"name"`          // 提供者名称
    Version      string    `json:"version"`       // 版本号
    Description  string    `json:"description"`   // 描述
    HealthStatus string    `json:"health_status"` // 健康状态
    LastCheck    time.Time `json:"last_check"`    // 最后检查时间
    MetricsCount int       `json:"metrics_count"` // 指标数量
}
```

### BaseProvider 结构体

```go
type BaseProvider struct {
    metadata ProviderMetadata
}
```

提供MetricProvider接口的默认实现。

#### NewBaseProvider
```go
func NewBaseProvider(name, version, description string) *BaseProvider
```
创建基础提供者。

#### 默认实现方法
```go
func (p *BaseProvider) Name() string
func (p *BaseProvider) Metrics() []MetricDefinition
func (p *BaseProvider) Init(collector interface{}) *types.Error
func (p *BaseProvider) Collect() ([]Metric, *types.Error)
func (p *BaseProvider) Shutdown() *types.Error
func (p *BaseProvider) GetMetadata() ProviderMetadata
```

## 🎯 指标收集器

### Collector 结构体

```go
type Collector struct {
    // 内部字段
}
```

### 构造函数

#### NewCollector
```go
func NewCollector(config *MonitoringConfig) (*Collector, *types.Error)
```
创建新的指标收集器。

### 方法

#### RegisterDefinition
```go
func (c *Collector) RegisterDefinition(def MetricDefinition) *types.Error
```
注册指标定义。

#### Record
```go
func (c *Collector) Record(metric Metric) *types.Error
```
记录指标值。

#### GetMetrics
```go
func (c *Collector) GetMetrics(name string, labels map[string]string) ([]Metric, *types.Error)
```
获取指定名称和标签的指标。

#### RegisterProvider
```go
func (c *Collector) RegisterProvider(name string, provider MetricProvider) *types.Error
```
注册指标提供者。

#### CollectFromProviders
```go
func (c *Collector) CollectFromProviders() *types.Error
```
从所有提供者收集指标。

#### GetStats
```go
func (c *Collector) GetStats() CollectorStats
```
获取收集器统计信息。

### CollectorStats 结构体

```go
type CollectorStats struct {
    MetricsCollected int64         `json:"metrics_collected"` // 收集的指标数量
    MetricsDropped   int64         `json:"metrics_dropped"`   // 丢弃的指标数量
    BatchOperations  int64         `json:"batch_operations"`  // 批量操作次数
    LastCollection   time.Time     `json:"last_collection"`   // 最后收集时间
    Uptime           time.Duration `json:"uptime"`            // 运行时间
    StartTime        time.Time     `json:"start_time"`        // 启动时间
}
```

## 📋 注册表管理

### Registry 结构体

```go
type Registry struct {
    // 内部字段
}
```

### 构造函数

#### NewRegistry
```go
func NewRegistry() *Registry
```
创建新的注册表。

### 方法

#### RegisterProviderFactory
```go
func (r *Registry) RegisterProviderFactory(
    name string,
    factory ProviderFactory,
    autoInit bool,
    priority int,
) *types.Error
```
注册提供者工厂。

#### InitializeProviders
```go
func (r *Registry) InitializeProviders(collector interface{}) *types.Error
```
初始化所有提供者。

#### GetProvider
```go
func (r *Registry) GetProvider(name string) (MetricProvider, *types.Error)
```
获取指定名称的提供者。

#### GetAllProviders
```go
func (r *Registry) GetAllProviders() []MetricProvider
```
获取所有提供者。

#### ShutdownProviders
```go
func (r *Registry) ShutdownProviders() *types.Error
```
关闭所有提供者。

## 📤 指标导出器

### Exporter 结构体

```go
type Exporter struct {
    // 内部字段
}
```

### 构造函数

#### NewExporter
```go
func NewExporter(collector *Collector, config *ExportConfig) (*Exporter, *types.Error)
```
创建新的指标导出器。

### 方法

#### Start
```go
func (e *Exporter) Start() *types.Error
```
启动HTTP服务器。

#### Stop
```go
func (e *Exporter) Stop() *types.Error
```
停止HTTP服务器。

#### ExportPrometheus
```go
func (e *Exporter) ExportPrometheus() (string, *types.Error)
```
导出为Prometheus格式。

#### ExportJSON
```go
func (e *Exporter) ExportJSON() (string, *types.Error)
```
导出为JSON格式。

#### GetStats
```go
func (e *Exporter) GetStats() ExporterStats
```
获取导出器统计信息。

### ExporterStats 结构体

```go
type ExporterStats struct {
    RequestsTotal  int64            `json:"requests_total"`   // 总请求数
    RequestsByPath map[string]int64 `json:"requests_by_path"` // 按路径统计的请求数
    ErrorsTotal    int64            `json:"errors_total"`     // 错误总数
    LastRequest    time.Time        `json:"last_request"`     // 最后请求时间
    StartTime      time.Time        `json:"start_time"`       // 启动时间
    Uptime         time.Duration    `json:"uptime"`           // 运行时间
    CacheHits      int64            `json:"cache_hits"`       // 缓存命中数
    CacheMisses    int64            `json:"cache_misses"`     // 缓存未命中数
}
```

## ❌ 错误处理

### Error 结构体

```go
type Error struct {
    Code    Code   `json:"code"`    // 错误代码
    Message string `json:"message"` // 错误消息
}
```

### Code 类型

```go
type Code int
```

### 错误代码常量

```go
const (
    MetricAlreadyRegistered    Code = 1000 + iota // 指标已注册
    MetricNotFound                                // 指标未找到
    InvalidMetricType                             // 无效的指标类型
    InvalidMetricValue                            // 无效的指标值
    CollectorNotInitialized                       // 收集器未初始化
    RegistryNotInitialized                        // 注册表未初始化
    ProviderAlreadyRegistered                     // 提供者已注册
    ProviderNotFound                              // 提供者未找到
    InvalidConfiguration                          // 无效的配置
    ExportFailed                                  // 导出失败
    SamplingDisabled                              // 采样已禁用
    BufferFull                                    // 缓冲区已满
    OperationTimeout                              // 操作超时
    ResourceExhausted                             // 资源耗尽
)
```

### 错误创建函数

#### NewError
```go
func NewError(code Code, message string) *Error
```
创建新的错误。

#### NewMetricAlreadyRegisteredError
```go
func NewMetricAlreadyRegisteredError(metricName string) *Error
```
创建指标已注册错误。

#### NewMetricNotFoundError
```go
func NewMetricNotFoundError(metricName string) *Error
```
创建指标未找到错误。

#### NewInvalidMetricTypeError
```go
func NewInvalidMetricTypeError(metricType string) *Error
```
创建无效指标类型错误。

#### NewCollectorNotInitializedError
```go
func NewCollectorNotInitializedError() *Error
```
创建收集器未初始化错误。

#### NewProviderAlreadyRegisteredError
```go
func NewProviderAlreadyRegisteredError(providerName string) *Error
```
创建提供者已注册错误。

#### NewInvalidConfigurationError
```go
func NewInvalidConfigurationError(field string, reason string) *Error
```
创建无效配置错误。

### 错误检查函数

#### IsMetricAlreadyRegistered
```go
func IsMetricAlreadyRegistered(err *Error) bool
```
检查是否为指标已注册错误。

#### IsMetricNotFound
```go
func IsMetricNotFound(err *Error) bool
```
检查是否为指标未找到错误。

#### IsInvalidConfiguration
```go
func IsInvalidConfiguration(err *Error) bool
```
检查是否为无效配置错误。

## 🌍 全局管理器

### 全局函数

#### InitializeGlobalManager
```go
func InitializeGlobalManager() *types.Error
```
初始化全局管理器。

#### ShutdownGlobalManager
```go
func ShutdownGlobalManager() *types.Error
```
关闭全局管理器。

#### RegisterGlobalProvider
```go
func RegisterGlobalProvider(
    name string,
    factory types.ProviderFactory,
    autoInit bool,
    priority int,
) *types.Error
```
注册全局提供者。

#### GetGlobalManager
```go
func GetGlobalManager() *Manager
```
获取全局管理器实例。

### 使用示例

```go
// 初始化全局管理器
monitoring.InitializeGlobalManager()
defer monitoring.ShutdownGlobalManager()

// 注册全局提供者
monitoring.RegisterGlobalProvider(
    "myapp",
    func() types.MetricProvider { return NewMyAppProvider() },
    true,
    100,
)

// 获取全局管理器并启动
manager := monitoring.GetGlobalManager()
if err := manager.Start(); err != nil {
    log.Fatal(err)
}
```

## 📝 代码示例

### 完整示例

```go
package main

import (
    "log"
    "time"

    "github.com/muidea/magicCommon/monitoring"
    "github.com/muidea/magicCommon/monitoring/core"
    "github.com/muidea/magicCommon/monitoring/types"
)

func main() {
    // 1. 创建配置
    config := core.ProductionConfig()
    config.Namespace = "myapp"
    
    // 2. 创建管理器
    manager, err := monitoring.NewManager(&config)
    if err != nil {
        log.Fatal(err)
    }
    
    // 3. 注册提供者
    err = manager.RegisterProvider(
        "demo",
        func() types.MetricProvider { return NewDemoProvider() },
        true,
        100,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // 4. 启动监控
    if err := manager.Start(); err != nil {
        log.Fatal(err)
    }
    defer manager.Shutdown()
    
    log.Println("Monitoring started on port 9090")
    time.Sleep(10 * time.Minute)
}

type DemoProvider struct {
    *types.BaseProvider
    count int64
}

func NewDemoProvider() *DemoProvider {
    return &DemoProvider{
        BaseProvider: types.NewBaseProvider("demo", "1.0.0", "Demo metrics"),
    }
}

func (p *DemoProvider) Metrics() []types.MetricDefinition {
    return []types.MetricDefinition{
        types.NewCounterDefinition(
            "demo_requests_total",
            "Total demo requests",
            []string{"type"},
            nil,
        ),
    }
}

func (p *DemoProvider) Collect() ([]types.Metric, *types.Error) {
    p.count++
    return []types.Metric{
        types.NewCounter(
            "demo_requests_total",
            float64(p.count),
            map[string]string{"type": "test"},
        ),
    }, nil
}
```

## 🔗 相关文档

- [设计文档](MONITORING_FRAMEWORK_DESIGN.md) - 详细架构设计
- [快速开始指南](QUICK_START.md) - 快速上手教程
- [最佳实践指南](BEST_PRACTICES.md) - 生产环境部署建议
- [测试示例](test/) - 完整的测试用例

---

*API参考文档版本: 1.0.0*
*最后更新: 2026-02-02*
*包版本: monitoring v1.0.0*