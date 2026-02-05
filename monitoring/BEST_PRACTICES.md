# 最佳实践指南

本文档提供magicCommon监控框架在生产环境中的最佳实践和建议。

## 🎯 目标读者

- 生产环境部署工程师
- 系统架构师
- 开发团队负责人
- 运维工程师

## 📋 目录

- [配置管理](#配置管理)
- [性能优化](#性能优化)
- [安全实践](#安全实践)
- [监控策略](#监控策略)
- [故障处理](#故障处理)
- [扩展性设计](#扩展性设计)
- [团队协作](#团队协作)

## ⚙️ 配置管理

### 环境特定配置

#### 开发环境
```go
config := core.DevelopmentConfig()
// 特点：
// - 采样率: 10%（减少性能影响）
// - 导出: 禁用（避免端口冲突）
// - 异步收集: 禁用（便于调试）
// - 详细级别: basic（基本指标）
```

#### 生产环境
```go
config := core.ProductionConfig()
// 特点：
// - 采样率: 50%（平衡性能和数据完整性）
// - 导出: 启用（带认证和TLS）
// - 异步收集: 启用（不阻塞业务逻辑）
// - 详细级别: standard（标准操作指标）
// - 安全: 启用认证和TLS
```

#### 高负载环境
```go
config := core.HighLoadConfig()
// 特点：
// - 采样率: 10%（最小化性能影响）
// - 批量大小: 1000（优化吞吐量）
// - 刷新间隔: 60s（减少导出频率）
// - 缓冲区: 5000（处理突发流量）
```

### 配置管理建议

#### 1. 使用环境变量
```go
func loadConfig() core.MonitoringConfig {
    config := core.ProductionConfig()
    
    // 从环境变量覆盖配置
    if port := os.Getenv("MONITORING_PORT"); port != "" {
        if p, err := strconv.Atoi(port); err == nil {
            config.ExportConfig.Port = p
        }
    }
    
    if namespace := os.Getenv("MONITORING_NAMESPACE"); namespace != "" {
        config.Namespace = namespace
    }
    
    if samplingRate := os.Getenv("MONITORING_SAMPLING_RATE"); samplingRate != "" {
        if rate, err := strconv.ParseFloat(samplingRate, 64); err == nil {
            config.SamplingRate = rate
        }
    }
    
    return config
}
```

#### 2. 配置验证
```go
func validateAndApplyConfig(config core.MonitoringConfig) (*monitoring.Manager, error) {
    // 验证配置
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %v", err)
    }
    
    // 创建管理器
    manager, err := monitoring.NewManager(&config)
    if err != nil {
        return nil, fmt.Errorf("failed to create manager: %v", err)
    }
    
    // 记录配置信息（避免记录敏感信息）
    log.Printf("Monitoring configured: namespace=%s, sampling=%.2f, port=%d",
        config.Namespace, config.SamplingRate, config.ExportConfig.Port)
    
    return manager, nil
}
```

#### 3. 配置热重载
```go
type ConfigManager struct {
    currentConfig core.MonitoringConfig
    manager       *monitoring.Manager
    mu            sync.RWMutex
}

func (cm *ConfigManager) ReloadConfig(newConfig core.MonitoringConfig) error {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    
    // 验证新配置
    if err := newConfig.Validate(); err != nil {
        return err
    }
    
    // 创建新管理器
    newManager, err := monitoring.NewManager(&newConfig)
    if err != nil {
        return err
    }
    
    // 启动新管理器
    if err := newManager.Start(); err != nil {
        return err
    }
    
    // 停止旧管理器
    if cm.manager != nil {
        cm.manager.Shutdown()
    }
    
    // 更新配置和管理器
    cm.currentConfig = newConfig
    cm.manager = newManager
    
    log.Println("Monitoring configuration reloaded successfully")
    return nil
}
```

## 🚀 性能优化

### 采样策略

#### 1. 分层采样
```go
func getSamplingRate(metricName string, environment string) float64 {
    // 关键指标：100%采样
    criticalMetrics := map[string]bool{
        "app_errors_total":     true,
        "app_requests_total":   true,
        "app_response_time":    true,
    }
    
    if criticalMetrics[metricName] {
        return 1.0
    }
    
    // 根据环境调整采样率
    switch environment {
    case "development":
        return 0.1  // 10%
    case "production":
        return 0.5  // 50%
    case "highload":
        return 0.1  // 10%
    default:
        return 0.3  // 30%
    }
}
```

#### 2. 动态采样
```go
type DynamicSamplingProvider struct {
    *types.BaseProvider
    currentLoad float64 // 0.0-1.0
}

func (p *DynamicSamplingProvider) Collect() ([]types.Metric, *types.Error) {
    // 根据当前负载调整采样
    samplingRate := 1.0 - p.currentLoad*0.8  // 负载越高，采样率越低
    
    return []types.Metric{
        types.NewGauge(
            "app_sampling_rate",
            samplingRate,
            map[string]string{"strategy": "dynamic"},
        ),
    }, nil
}
```

### 内存管理

#### 1. 缓冲区大小调整
```go
func calculateBufferSize(expectedQPS int, retentionSeconds int) int {
    // 缓冲区大小 = QPS * 保留时间 * 安全系数
    bufferSize := expectedQPS * retentionSeconds * 2
    
    // 限制最小和最大值
    if bufferSize < 100 {
        return 100
    }
    if bufferSize > 10000 {
        return 10000
    }
    
    return bufferSize
}
```

#### 2. 指标保留策略
```go
config := core.ProductionConfig()

// 根据指标类型设置不同的保留时间
config.RetentionPeriod = 24 * time.Hour  // 默认保留24小时

// 高频指标：短期保留
if strings.HasPrefix(metricName, "app_requests_") {
    retention = 1 * time.Hour
}

// 低频指标：长期保留  
if strings.HasPrefix(metricName, "app_business_") {
    retention = 7 * 24 * time.Hour
}
```

### 并发优化

#### 1. 并发任务数
```go
func calculateConcurrentTasks(cpuCores int, memoryGB int) int {
    // 基础并发数 = CPU核心数 * 2
    baseTasks := cpuCores * 2
    
    // 根据内存调整
    memoryTasks := memoryGB * 10
    
    // 取较小值
    if baseTasks < memoryTasks {
        return baseTasks
    }
    return memoryTasks
}
```

#### 2. 批量处理优化
```go
type OptimizedCollector struct {
    batchSize   int
    batchBuffer []types.Metric
    flushTicker *time.Ticker
}

func (oc *OptimizedCollector) recordMetric(metric types.Metric) {
    oc.batchBuffer = append(oc.batchBuffer, metric)
    
    // 达到批量大小时立即刷新
    if len(oc.batchBuffer) >= oc.batchSize {
        oc.flushBatch()
    }
}

func (oc *OptimizedCollector) flushBatch() {
    if len(oc.batchBuffer) == 0 {
        return
    }
    
    // 批量处理逻辑
    processBatch(oc.batchBuffer)
    
    // 清空缓冲区
    oc.batchBuffer = oc.batchBuffer[:0]
}
```

## 🔒 安全实践

### TLS配置

#### 1. 自动证书管理
```go
func setupTLS(config *core.ExportConfig) error {
    if !config.EnableTLS {
        return nil
    }
    
    // 检查证书文件是否存在
    if _, err := os.Stat(config.TLSCertPath); os.IsNotExist(err) {
        // 自动生成自签名证书（仅用于开发）
        if config.Environment == "development" {
            cert, key, err := generateSelfSignedCert()
            if err != nil {
                return err
            }
            
            if err := os.WriteFile(config.TLSCertPath, cert, 0600); err != nil {
                return err
            }
            if err := os.WriteFile(config.TLSKeyPath, key, 0600); err != nil {
                return err
            }
            
            log.Println("Generated self-signed TLS certificate for development")
        } else {
            return fmt.Errorf("TLS certificate not found: %s", config.TLSCertPath)
        }
    }
    
    return nil
}
```

#### 2. 安全协议配置
```go
func createSecureServer(config *core.ExportConfig, handler http.Handler) *http.Server {
    tlsConfig := &tls.Config{
        MinVersion: tls.VersionTLS12, // 最低TLS 1.2
        CipherSuites: []uint16{
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
            tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
        },
        CurvePreferences: []tls.CurveID{
            tls.CurveP256,
            tls.CurveP384,
            tls.X25519,
        },
        PreferServerCipherSuites: true,
    }
    
    return &http.Server{
        Addr:         fmt.Sprintf(":%d", config.Port),
        Handler:      handler,
        TLSConfig:    tlsConfig,
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  120 * time.Second,
    }
}
```

### 访问控制

#### 1. IP白名单
```go
type IPWhitelistMiddleware struct {
    allowedIPs map[string]bool
    cidrBlocks []*net.IPNet
}

func (ipwl *IPWhitelistMiddleware) Handler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        clientIP := getClientIP(r)
        
        // 检查精确IP匹配
        if ipwl.allowedIPs[clientIP] {
            next.ServeHTTP(w, r)
            return
        }
        
        // 检查CIDR块匹配
        ip := net.ParseIP(clientIP)
        for _, cidr := range ipwl.cidrBlocks {
            if cidr.Contains(ip) {
                next.ServeHTTP(w, r)
                return
            }
        }
        
        // 记录未授权访问
        log.Printf("Unauthorized access attempt from IP: %s", clientIP)
        http.Error(w, "Forbidden", http.StatusForbidden)
    })
}
```

#### 2. 速率限制
```go
type RateLimitMiddleware struct {
    limiter *rate.Limiter
}

func NewRateLimitMiddleware(rps int) *RateLimitMiddleware {
    return &RateLimitMiddleware{
        limiter: rate.NewLimiter(rate.Limit(rps), rps*2),
    }
}

func (rlm *RateLimitMiddleware) Handler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !rlm.limiter.Allow() {
            w.Header().Set("Retry-After", "1")
            http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

## 📊 监控策略

### 指标设计

#### 1. 命名规范
```go
// 好的命名
"app_http_requests_total"
"app_database_query_duration_seconds"
"app_cache_hit_ratio"

// 避免的命名
"requests"                    // 缺少前缀
"db_query_time"              // 单位不明确
"cachehitratio"              // 缺少下划线分隔
```

#### 2. 标签设计原则
```go
// 好的标签设计
labels := map[string]string{
    "method":      "GET",      // 低基数
    "status_code": "200",      // 有限值
    "endpoint":    "/api/users", // 稳定值
}

// 避免的标签设计
labels := map[string]string{
    "user_id":     "12345",    // 高基数 - 避免！
    "session_id":  "abc123",   // 高基数 - 避免！
    "timestamp":   "2025-01-01T12:00:00Z", // 不断变化 - 避免！
}
```

#### 3. 指标层次结构
```go
// 系统级指标
"app_cpu_usage_percent"
"app_memory_usage_bytes"
"app_goroutine_count"

// 应用级指标
"app_http_requests_total"
"app_http_request_duration_seconds"
"app_http_errors_total"

// 业务级指标
"app_orders_created_total"
"app_payments_processed_total"
"app_users_active_count"

// 依赖服务指标
"app_database_connections_active"
"app_cache_hit_ratio"
"app_external_api_call_duration_seconds"
```

### 告警策略

#### 1. 告警规则定义
```go
type AlertRule struct {
    MetricName    string
    Condition     string  // ">", "<", "==", "!="
    Threshold     float64
    Duration      time.Duration
    Severity      string  // "critical", "warning", "info"
    NotifyChannels []string
}

var alertRules = []AlertRule{
    {
        MetricName:    "app_http_errors_ratio",
        Condition:     ">",
        Threshold:     0.05, // 5%错误率
        Duration:      5 * time.Minute,
        Severity:      "critical",
        NotifyChannels: []string{"slack", "pagerduty"},
    },
    {
        MetricName:    "app_response_time_p95",
        Condition:     ">",
        Threshold:     2.0, // 2秒
        Duration:      10 * time.Minute,
        Severity:      "warning",
        NotifyChannels: []string{"slack"},
    },
}
```

#### 2. 告警抑制
```go
type AlertSuppression struct {
    MatchingLabels map[string]string
    Duration       time.Duration
    Reason         string
}

var alertSuppressions = []AlertSuppression{
    {
        MatchingLabels: map[string]string{
            "environment": "staging",
        },
        Duration: 0, // 永久抑制
        Reason:   "Staging environment, no alerts needed",
    },
    {
        MatchingLabels: map[string]string{
            "maintenance": "true",
        },
        Duration: 2 * time.Hour,
        Reason:   "Scheduled maintenance window",
    },
}
```

## 🛠️ 故障处理

### 故障检测

#### 1. 健康检查
```go
type HealthChecker struct {
    manager *monitoring.Manager
}

func (hc *HealthChecker) Check() HealthStatus {
    status := HealthStatus{
        Status:  "healthy",
        Checks:  make(map[string]string),
        Details: make(map[string]interface{}),
    }
    
    // 检查管理器状态
    if hc.manager == nil {
        status.Status = "unhealthy"
        status.Checks["manager"] = "not initialized"
        return status
    }
    
    // 检查提供者健康状态
    providers := hc.manager.GetAllProviders()
    for _, provider := range providers {
        metadata := provider.GetMetadata()
        if metadata.HealthStatus != "healthy" {
            status.Status = "degraded"
            status.Checks[provider.Name()] = metadata.HealthStatus
        }
    }
    
    // 检查收集器缓冲区
    stats := hc.manager.GetCollectorStats()
    if stats.BufferUsage > 0.9 { // 90%使用率
        status.Status = "degraded"
        status.Checks["buffer"] = "high usage"
        status.Details["buffer_usage"] = stats.BufferUsage
    }
    
    return status
}
```

#### 2. 自动恢复
```go
type AutoRecovery struct {
    manager        *monitoring.Manager
    failureCount   int
    lastFailure    time.Time
    recoveryTicker *time.Ticker
}

func (ar *AutoRecovery) Monitor() {
    for {
        select {
        case <-ar.recoveryTicker.C:
            if !ar.isHealthy() {
                ar.failureCount++
                
                // 尝试恢复
                if ar.failureCount >= 3 {
                    ar.attemptRecovery()
                }
            } else {
                ar.failureCount = 0
            }
        }
    }
}

func (ar *AutoRecovery) attemptRecovery() {
    log.Println("Attempting automatic recovery...")
    
    // 1. 尝试重启管理器
    ar.manager.Shutdown()
    time.Sleep(1 * time.Second)
    
    if err := ar.manager.Start(); err != nil {
        log.Printf("Recovery failed: %v", err)
        
        // 2. 回退到安全模式
        ar.enterSafeMode()
    } else {
        log.Println("Recovery successful")
        ar.failureCount = 0
    }
}
```

### 故障排查流程

#### 1. 诊断检查清单
```markdown
# 监控故障排查清单

## 症状：指标未收集
- [ ] 检查监控是否启用 (`config.Enabled`)
- [ ] 验证采样率配置 (`config.SamplingRate > 0`)
- [ ] 检查提供者注册状态
- [ ] 查看收集器日志
- [ ] 验证指标定义是否正确注册

## 症状：内存使用过高
- [ ] 检查缓冲区大小 (`config.BufferSize`)
- [ ] 查看指标保留时间 (`config.RetentionPeriod`)
- [ ] 检查是否有内存泄漏
- [ ] 验证采样率是否过低

## 症状：导出失败
- [ ] 检查端口占用 (`config.ExportConfig.Port`)
- [ ] 验证认证配置
- [ ] 检查网络连接
- [ ] 查看导出器日志
```

#### 2. 调试工具
```go
func enableDebugMode(config *core.MonitoringConfig) {
    // 启用详细日志
    config.DetailLevel = core.DetailLevelDetailed
    
    // 降低采样率便于调试
    config.SamplingRate = 1.0
    
    // 禁用异步收集
    config.AsyncCollection = false
    
    // 启用pprof
    config.ExportConfig.EnablePProf = true
    
    log.Println("Debug mode enabled")
}
```

## 🏗️ 扩展性设计

### 自定义提供者模式

#### 1. 工厂模式
```go
type ProviderFactoryRegistry struct {
    factories map[string]types.ProviderFactory
}

func (pfr *ProviderFactoryRegistry) Register(name string, factory types.ProviderFactory) {
    pfr.factories[name] = factory
}

func (pfr *ProviderFactoryRegistry) CreateProvider(name string, config map[string]interface{}) (types.MetricProvider, error) {
    factory, exists := pfr.factories[name]
    if !exists {
        return nil, fmt.Errorf("provider factory not found: %s", name)
    }
    
    provider := factory()
    
    // 应用配置
    if configProvider, ok := provider.(ConfigurableProvider); ok {
        if err := configProvider.Configure(config); err != nil {
            return nil, err
        }
    }
    
    return provider, nil
}
```

#### 2. 装饰器模式
```go
type CachingProvider struct {
    provider types.MetricProvider
    cache    map[string][]types.Metric
    cacheTTL time.Duration
    lastUpdate time.Time
}

func (cp *CachingProvider) Collect() ([]types.Metric, *types.Error) {
    // 检查缓存是否有效
    if time.Since(cp.lastUpdate) < cp.cacheTTL && cp.cache != nil {
        // 返回缓存数据
        return cp.flattenCache(), nil
    }
    
    // 从底层提供者收集
    metrics, err := cp.provider.Collect()
    if err != nil {
        return nil, err
    }
    
    // 更新缓存
    cp.updateCache(metrics)
    cp.lastUpdate = time.Now()
    
    return metrics, nil
}

func (cp *CachingProvider) updateCache(metrics []types.Metric) {
    cp.cache = make(map[string][]types.Metric)
    for _, metric := range metrics {
        key := metric.Name
        cp.cache[key] = append(cp.cache[key], metric)
    }
}
```

### 分布式监控

#### 1. 聚合模式
```go
type AggregatingProvider struct {
    *types.BaseProvider
    childProviders []types.MetricProvider
    aggregationRules map[string]AggregationRule
}

type AggregationRule struct {
    Operation string // "sum", "avg", "min", "max"
    GroupBy   []string
}

func (ap *AggregatingProvider) Collect() ([]types.Metric, *types.Error) {
    allMetrics := make([]types.Metric, 0)
    
    // 从所有子提供者收集
    for _, provider := range ap.childProviders {
        metrics, err := provider.Collect()
        if err != nil {
            continue // 跳过失败的提供者
        }
        allMetrics = append(allMetrics, metrics...)
    }
    
    // 应用聚合规则
    aggregatedMetrics := ap.aggregateMetrics(allMetrics)
    
    return aggregatedMetrics, nil
}
```

#### 2. 联邦模式
```go
type FederatedExporter struct {
    upstreamEndpoints []string
    authToken        string
    cache            *ttlcache.Cache
}

func (fe *FederatedExporter) Export() (string, error) {
    var allMetrics []types.Metric
    
    // 并行从所有上游收集
    var wg sync.WaitGroup
    var mu sync.Mutex
    
    for _, endpoint := range fe.upstreamEndpoints {
        wg.Add(1)
        go func(ep string) {
            defer wg.Done()
            
            metrics, err := fe.collectFromEndpoint(ep)
            if err != nil {
                log.Printf("Failed to collect from %s: %v", ep, err)
                return
            }
            
            mu.Lock()
            allMetrics = append(allMetrics, metrics...)
            mu.Unlock()
        }(endpoint)
    }
    
    wg.Wait()
    
    // 格式化为Prometheus
    return fe.formatAsPrometheus(allMetrics), nil
}
```

## 👥 团队协作

### 开发流程

#### 1. 代码审查清单
```markdown
# 监控代码审查清单

## 指标设计
- [ ] 指标名称符合命名规范
- [ ] 标签设计合理（避免高基数）
- [ ] 指标类型选择正确
- [ ] 帮助文本清晰明确

## 性能考虑
- [ ] 采样率设置合理
- [ ] 避免在热路径中收集指标
- [ ] 批量处理优化
- [ ] 内存使用可控

## 错误处理
- [ ] 所有错误都得到处理
- [ ] 错误信息清晰可读
- [ ] 有适当的降级策略

## 测试覆盖
- [ ] 单元测试覆盖核心逻辑
- [ ] 集成测试验证端到端功能
- [ ] 性能测试验证资源使用
```

#### 2. 文档要求
```go
// Good documentation example
type RequestMetricsProvider struct {
    *types.BaseProvider
    
    // requestCount tracks the total number of HTTP requests
    // This is a counter that only increases
    requestCount int64
    
    // activeConnections tracks current active connections
    // This is a gauge that can go up and down
    activeConnections int32
}

// Metrics returns the metric definitions for this provider.
// It defines two metrics:
// - app_http_requests_total: Total HTTP requests processed
// - app_http_active_connections: Current active HTTP connections
func (p *RequestMetricsProvider) Metrics() []types.MetricDefinition {
    return []types.MetricDefinition{
        types.NewCounterDefinition(
            "app_http_requests_total",
            "Total number of HTTP requests processed",
            []string{"method", "status_code", "endpoint"},
            map[string]string{"service": "api"},
        ),
        types.NewGaugeDefinition(
            "app_http_active_connections", 
            "Current number of active HTTP connections",
            []string{"protocol"}, // "http", "https"
            nil,
        ),
    }
}
```

### 运维手册

#### 1. 部署检查清单
```markdown
# 生产部署检查清单

## 预部署检查
- [ ] 配置验证通过
- [ ] 安全审查完成
- [ ] 性能测试通过
- [ ] 回滚计划准备

## 部署过程
- [ ] 备份当前配置
- [ ] 逐步部署（金丝雀发布）
- [ ] 监控部署过程
- [ ] 验证功能正常

## 部署后检查
- [ ] 指标收集正常
- [ ] 导出功能正常
- [ ] 性能指标在预期范围内
- [ ] 告警规则生效
```

#### 2. 运维命令
```bash
# 检查监控状态
curl http://localhost:9090/health

# 查看当前指标
curl http://localhost:9090/metrics

# 查看系统信息
curl http://localhost:9090/

# 性能分析
go tool pprof http://localhost:9090/debug/pprof/heap

# 日志查看
tail -f /var/log/monitoring.log

# 配置重载
kill -HUP $(pidof myapp)
```

## 📚 总结

### 关键要点

1. **配置管理**: 使用环境特定配置，支持热重载
2. **性能优化**: 合理设置采样率，优化内存使用
3. **安全实践**: 启用认证和TLS，实施访问控制
4. **监控策略**: 设计合理的指标和告警规则
5. **故障处理**: 建立完善的诊断和恢复流程
6. **扩展性**: 支持自定义提供者和分布式部署
7. **团队协作**: 建立代码审查和运维流程

### 持续改进

监控系统需要持续改进和优化：

1. **定期审查**: 每月审查指标设计和告警规则
2. **性能调优**: 根据实际负载调整配置参数
3. **安全更新**: 定期更新安全配置和证书
4. **功能扩展**: 根据业务需求添加新的监控功能
5. **文档更新**: 保持文档与代码同步更新

### 资源推荐

- [Prometheus最佳实践](https://prometheus.io/docs/practices/naming/)
- [监控模式](https://landing.google.com/sre/sre-book/chapters/monitoring-distributed-systems/)
- [可观测性工程](https://www.oreilly.com/library/view/observability-engineering/9781492076438/)

---

*最佳实践指南版本: 1.0.0*
*最后更新: 2026-02-02*
*适用环境: 生产环境*