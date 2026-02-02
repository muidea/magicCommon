package core

import (
	"sync"
	"time"

	"github.com/muidea/magicCommon/monitoring/types"
)

// Collector collects and manages metrics
type Collector struct {
	mu sync.RWMutex

	config      *MonitoringConfig
	metrics     map[string][]types.Metric
	definitions map[string]types.MetricDefinition

	// Performance optimization
	batchBuffer []types.Metric
	batchSize   int
	batchMutex  sync.Mutex

	// Statistics
	stats CollectorStats

	// Provider management
	providers map[string]types.MetricProvider

	// Background task control
	stopChan chan struct{}
}

// CollectorStats holds collector statistics
type CollectorStats struct {
	MetricsCollected int64         `json:"metrics_collected"`
	MetricsDropped   int64         `json:"metrics_dropped"`
	BatchOperations  int64         `json:"batch_operations"`
	LastCollection   time.Time     `json:"last_collection"`
	Uptime           time.Duration `json:"uptime"`
	StartTime        time.Time     `json:"start_time"`
}

// NewCollector creates a new metric collector
func NewCollector(config *MonitoringConfig) (*Collector, *types.Error) {
	if config == nil {
		defaultConfig := DefaultMonitoringConfig()
		config = &defaultConfig
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	collector := &Collector{
		config:      config,
		metrics:     make(map[string][]types.Metric),
		definitions: make(map[string]types.MetricDefinition),
		batchBuffer: make([]types.Metric, 0, config.BufferSize),
		batchSize:   config.BatchSize,
		providers:   make(map[string]types.MetricProvider),
		stats: CollectorStats{
			StartTime: time.Now(),
		},
	}

	// Register default metric definitions
	collector.registerDefaultDefinitions()

	// Start background tasks if async collection is enabled
	if config.AsyncCollection {
		go collector.startBackgroundTasks()
	}

	return collector, nil
}

// RegisterDefinition registers a new metric definition
func (c *Collector) RegisterDefinition(def types.MetricDefinition) *types.Error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Add namespace prefix
	fullName := def.GetFullName(c.config.Namespace)
	if _, exists := c.definitions[fullName]; exists {
		return types.NewMetricAlreadyRegisteredError(fullName)
	}

	// Validate definition
	if err := def.Validate(); err != nil {
		return err
	}

	// Store with full name
	def.Name = fullName
	c.definitions[fullName] = def
	return nil
}

// Record records a metric value
func (c *Collector) Record(name string, value float64, labels map[string]string) *types.Error {
	if !c.config.ShouldSample() {
		c.stats.MetricsDropped++
		return nil
	}

	def, err := c.getDefinition(name)
	if err != nil {
		return err
	}

	// Validate labels
	if err := c.validateLabels(def, labels); err != nil {
		return err
	}

	metric := types.NewMetric(name, def.Type, value, labels)

	if c.config.AsyncCollection {
		return c.recordAsync(metric)
	}

	return c.recordSync(metric)
}

// RecordWithTimestamp records a metric with a specific timestamp
func (c *Collector) RecordWithTimestamp(name string, value float64, labels map[string]string, timestamp time.Time) *types.Error {
	if !c.config.ShouldSample() {
		c.stats.MetricsDropped++
		return nil
	}

	def, err := c.getDefinition(name)
	if err != nil {
		return err
	}

	// Validate labels
	if err := c.validateLabels(def, labels); err != nil {
		return err
	}

	metric := types.Metric{
		Name:      name,
		Type:      def.Type,
		Value:     value,
		Labels:    labels,
		Timestamp: timestamp,
	}

	if c.config.AsyncCollection {
		return c.recordAsync(metric)
	}

	return c.recordSync(metric)
}

// Increment increments a counter metric
func (c *Collector) Increment(name string, labels map[string]string) *types.Error {
	return c.Record(name, 1, labels)
}

// Decrement decrements a gauge metric
func (c *Collector) Decrement(name string, labels map[string]string) *types.Error {
	return c.Record(name, -1, labels)
}

// Observe observes a value for a histogram or summary metric
func (c *Collector) Observe(name string, value float64, labels map[string]string) *types.Error {
	return c.Record(name, value, labels)
}

// GetMetrics returns all collected metrics
func (c *Collector) GetMetrics() map[string][]types.Metric {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to avoid concurrent modification
	result := make(map[string][]types.Metric)
	for name, metrics := range c.metrics {
		result[name] = make([]types.Metric, len(metrics))
		copy(result[name], metrics)
	}
	return result
}

// GetMetricsByName returns metrics for a specific name
func (c *Collector) GetMetricsByName(name string) ([]types.Metric, *types.Error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	metrics, exists := c.metrics[name]
	if !exists {
		return nil, types.NewMetricNotFoundError(name)
	}

	// Return a copy
	result := make([]types.Metric, len(metrics))
	copy(result, metrics)
	return result, nil
}

// GetDefinitions returns all metric definitions
func (c *Collector) GetDefinitions() map[string]types.MetricDefinition {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy
	result := make(map[string]types.MetricDefinition)
	for name, def := range c.definitions {
		result[name] = def
	}
	return result
}

// GetDefinition returns a specific metric definition
func (c *Collector) GetDefinition(name string) (types.MetricDefinition, *types.Error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	def, exists := c.definitions[name]
	if !exists {
		return types.MetricDefinition{}, types.NewMetricNotFoundError(name)
	}
	return def, nil
}

// ClearMetrics clears all collected metrics
func (c *Collector) ClearMetrics() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics = make(map[string][]types.Metric)
	c.batchBuffer = make([]types.Metric, 0, c.config.BufferSize)
}

// ClearMetricsByName clears metrics for a specific name
func (c *Collector) ClearMetricsByName(name string) *types.Error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.metrics[name]; !exists {
		return types.NewMetricNotFoundError(name)
	}

	delete(c.metrics, name)
	return nil
}

// GetStats returns collector statistics
func (c *Collector) GetStats() CollectorStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.Uptime = time.Since(stats.StartTime)
	return stats
}

// RegisterProvider registers a metric provider
func (c *Collector) RegisterProvider(provider types.MetricProvider) *types.Error {
	name := provider.Name()

	// Check if provider already exists (with lock)
	c.mu.Lock()
	if _, exists := c.providers[name]; exists {
		c.mu.Unlock()
		return types.NewProviderAlreadyRegisteredError(name)
	}
	c.mu.Unlock()

	// Initialize the provider WITHOUT holding the lock
	if err := provider.Init(c); err != nil {
		return err
	}

	// Register provider's metric definitions
	for _, def := range provider.Metrics() {
		fullName := def.GetFullName(c.config.Namespace)
		def.Name = fullName
		if err := c.RegisterDefinition(def); err != nil {
			return err
		}
	}

	// Add provider to map (with lock)
	c.mu.Lock()
	c.providers[name] = provider
	c.mu.Unlock()

	return nil
}

// UnregisterProvider unregisters a metric provider
func (c *Collector) UnregisterProvider(name string) *types.Error {
	c.mu.Lock()
	defer c.mu.Unlock()

	provider, exists := c.providers[name]
	if !exists {
		return types.NewProviderNotFoundError(name)
	}

	// Shutdown the provider
	if err := provider.Shutdown(); err != nil {
		return err
	}

	delete(c.providers, name)
	return nil
}

// CollectFromProviders collects metrics from all registered providers
func (c *Collector) CollectFromProviders() *types.Error {
	c.mu.RLock()
	providers := make([]types.MetricProvider, 0, len(c.providers))
	for _, provider := range c.providers {
		providers = append(providers, provider)
	}
	c.mu.RUnlock()

	var lastError *types.Error
	for _, provider := range providers {
		startTime := time.Now()
		metrics, err := provider.Collect()
		duration := time.Since(startTime)

		if err != nil {
			lastError = err
			// Update provider stats with error
			if baseProvider, ok := provider.(interface{ UpdateLastError(*types.Error) }); ok {
				baseProvider.UpdateLastError(err)
			}
			continue
		}

		// Record collected metrics
		for _, metric := range metrics {
			// Ensure metric has the namespace prefix
			if c.config.Namespace != "" && metric.Name != "" {
				metric.Name = c.config.Namespace + "_" + metric.Name
			}

			if c.config.AsyncCollection {
				if err := c.recordAsync(metric); err != nil {
					lastError = err
				}
			} else {
				if err := c.recordSync(metric); err != nil {
					lastError = err
				}
			}
		}

		// Update provider stats
		if baseProvider, ok := provider.(interface {
			UpdateCollectionStats(bool, time.Duration, int)
		}); ok {
			baseProvider.UpdateCollectionStats(true, duration, len(metrics))
		}
	}

	return lastError
}

// GetProviders returns all registered providers
func (c *Collector) GetProviders() map[string]types.MetricProvider {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy
	result := make(map[string]types.MetricProvider)
	for name, provider := range c.providers {
		result[name] = provider
	}
	return result
}

// GetProvider returns a specific provider
func (c *Collector) GetProvider(name string) (types.MetricProvider, *types.Error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	provider, exists := c.providers[name]
	if !exists {
		return nil, types.NewProviderNotFoundError(name)
	}
	return provider, nil
}

// Private helper methods

func (c *Collector) getDefinition(name string) (types.MetricDefinition, *types.Error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	def, exists := c.definitions[name]
	if !exists {
		return types.MetricDefinition{}, types.NewMetricNotFoundError(name)
	}
	return def, nil
}

func (c *Collector) validateLabels(def types.MetricDefinition, labels map[string]string) *types.Error {
	// Check for required labels
	for _, labelName := range def.LabelNames {
		if _, exists := labels[labelName]; !exists {
			return types.NewInvalidConfigurationError("label_"+labelName, "", "required label missing")
		}
	}

	// Check for extra labels (only if we want to be strict)
	// For now, we allow extra labels

	return nil
}

func (c *Collector) recordSync(metric types.Metric) *types.Error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics[metric.Name] = append(c.metrics[metric.Name], metric)
	c.stats.MetricsCollected++
	c.stats.LastCollection = time.Now()

	// Check retention period
	c.cleanupOldMetrics()

	return nil
}

func (c *Collector) recordAsync(metric types.Metric) *types.Error {
	c.batchMutex.Lock()
	defer c.batchMutex.Unlock()

	c.batchBuffer = append(c.batchBuffer, metric)

	// Flush if buffer is full
	if len(c.batchBuffer) >= c.batchSize {
		return c.flushBatch()
	}

	return nil
}

func (c *Collector) flushBatch() *types.Error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, metric := range c.batchBuffer {
		c.metrics[metric.Name] = append(c.metrics[metric.Name], metric)
		c.stats.MetricsCollected++
	}

	c.stats.LastCollection = time.Now()
	c.stats.BatchOperations++
	c.batchBuffer = c.batchBuffer[:0]

	// Check retention period
	c.cleanupOldMetrics()

	return nil
}

func (c *Collector) cleanupOldMetrics() int {
	if c.config.RetentionPeriod <= 0 {
		return 0
	}

	cutoffTime := time.Now().Add(-c.config.RetentionPeriod)
	totalRemoved := 0

	for name, metrics := range c.metrics {
		// Filter out old metrics
		filtered := make([]types.Metric, 0, len(metrics))
		for _, metric := range metrics {
			if metric.Timestamp.After(cutoffTime) {
				filtered = append(filtered, metric)
			} else {
				totalRemoved++
			}
		}
		c.metrics[name] = filtered

		// Remove empty metric lists
		if len(filtered) == 0 {
			delete(c.metrics, name)
		}
	}

	return totalRemoved
}

func (c *Collector) startBackgroundTasks() {
	ticker := time.NewTicker(c.config.CollectionInterval)
	defer ticker.Stop()

	// Create a stop channel
	stopChan := make(chan struct{})

	// Store stop channel for shutdown
	c.mu.Lock()
	c.stopChan = stopChan
	c.mu.Unlock()

	for {
		select {
		case <-ticker.C:
			// Collect from providers
			c.CollectFromProviders()

			// Flush batch buffer
			c.batchMutex.Lock()
			if len(c.batchBuffer) > 0 {
				c.flushBatch()
			}
			c.batchMutex.Unlock()

			// Cleanup old metrics
			c.mu.Lock()
			c.cleanupOldMetrics()
			c.mu.Unlock()
		case <-stopChan:
			// Stop background tasks
			return
		}
	}
}

func (c *Collector) registerDefaultDefinitions() {
	// Register default system metrics
	defaultDefinitions := []types.MetricDefinition{
		types.NewCounterDefinition(
			"monitoring_metrics_collected_total",
			"Total number of metrics collected",
			[]string{},
			nil,
		),
		types.NewCounterDefinition(
			"monitoring_metrics_dropped_total",
			"Total number of metrics dropped due to sampling",
			[]string{},
			nil,
		),
		types.NewGaugeDefinition(
			"monitoring_collector_uptime_seconds",
			"Collector uptime in seconds",
			[]string{},
			nil,
		),
		types.NewGaugeDefinition(
			"monitoring_batch_buffer_size",
			"Current batch buffer size",
			[]string{},
			nil,
		),
	}

	for _, def := range defaultDefinitions {
		fullName := def.GetFullName(c.config.Namespace)
		def.Name = fullName
		c.definitions[def.Name] = def
	}
}

// Shutdown gracefully shuts down the collector
func (c *Collector) Shutdown() *types.Error {
	c.mu.Lock()

	// Stop background tasks if running
	if c.stopChan != nil {
		close(c.stopChan)
		c.stopChan = nil
	}

	// Shutdown all providers
	var lastError *types.Error
	for name, provider := range c.providers {
		if err := provider.Shutdown(); err != nil {
			lastError = err
		}
		delete(c.providers, name)
	}

	// Clear all metrics
	c.metrics = make(map[string][]types.Metric)
	c.batchBuffer = nil
	c.definitions = make(map[string]types.MetricDefinition)

	c.mu.Unlock()

	return lastError
}

// GetProviderMetrics returns metrics from a specific provider
func (c *Collector) GetProviderMetrics(providerName string) ([]types.Metric, *types.Error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	provider, exists := c.providers[providerName]
	if !exists {
		return nil, types.NewProviderNotFoundError(providerName)
	}

	return provider.Collect()
}

// CleanupExpiredMetrics removes metrics older than retention period
func (c *Collector) CleanupExpiredMetrics() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.cleanupOldMetrics()
}

// GetBufferUsage returns current buffer usage percentage
func (c *Collector) GetBufferUsage() float64 {
	c.batchMutex.Lock()
	defer c.batchMutex.Unlock()

	if c.config.BufferSize <= 0 {
		return 0
	}

	return float64(len(c.batchBuffer)) / float64(c.config.BufferSize)
}

// IsBufferFull checks if the buffer is full
func (c *Collector) IsBufferFull() bool {
	c.batchMutex.Lock()
	defer c.batchMutex.Unlock()

	return len(c.batchBuffer) >= c.config.BufferSize
}

// ForceFlush forces a flush of the batch buffer
func (c *Collector) ForceFlush() *types.Error {
	c.batchMutex.Lock()
	defer c.batchMutex.Unlock()

	if len(c.batchBuffer) == 0 {
		return nil
	}

	return c.flushBatch()
}

// GetMetricCount returns the total number of metrics stored
func (c *Collector) GetMetricCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := 0
	for _, metrics := range c.metrics {
		total += len(metrics)
	}
	return total
}

// GetProviderCount returns the number of registered providers
func (c *Collector) GetProviderCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.providers)
}
