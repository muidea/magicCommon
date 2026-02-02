package monitoring

import (
	"sync"
	"time"

	"github.com/muidea/magicCommon/monitoring/core"
	"github.com/muidea/magicCommon/monitoring/types"
)

// Manager is the main entry point for the monitoring system
type Manager struct {
	mu sync.RWMutex

	config    *core.MonitoringConfig
	collector *core.Collector
	registry  *core.Registry
	exporter  *core.Exporter

	// State
	initialized bool
	running     bool

	// Statistics
	stats ManagerStats
}

// ManagerStats holds manager statistics
type ManagerStats struct {
	StartTime        int64 `json:"start_time"`
	UptimeSeconds    int64 `json:"uptime_seconds"`
	TotalMetrics     int64 `json:"total_metrics"`
	ActiveProviders  int64 `json:"active_providers"`
	ExportRequests   int64 `json:"export_requests"`
	CollectionCycles int64 `json:"collection_cycles"`
}

// NewManager creates a new monitoring manager
func NewManager(config *core.MonitoringConfig) (*Manager, *types.Error) {
	if config == nil {
		defaultConfig := core.DefaultMonitoringConfig()
		config = &defaultConfig
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	manager := &Manager{
		config: config,
		stats: ManagerStats{
			StartTime: time.Now().Unix(),
		},
	}

	return manager, nil
}

// Initialize initializes the monitoring system
func (m *Manager) Initialize() *types.Error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.initializeLocked()
}

// initializeLocked performs initialization while holding the lock
// Caller must hold m.mu
func (m *Manager) initializeLocked() *types.Error {
	if m.initialized {
		return nil // Already initialized
	}

	// Create collector
	collector, err := core.NewCollector(m.config)
	if err != nil {
		return err
	}
	m.collector = collector

	// Create registry
	registry, err := core.NewRegistry(collector, m.config)
	if err != nil {
		return err
	}
	m.registry = registry

	// Create exporter if export is enabled
	if m.config.IsExportEnabled() {
		exporter, err := core.NewExporter(collector, &m.config.ExportConfig)
		if err != nil {
			return err
		}
		m.exporter = exporter
	}

	m.initialized = true
	return nil
}

// Start starts the monitoring system
func (m *Manager) Start() *types.Error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		if err := m.initializeLocked(); err != nil {
			return err
		}
	}

	if m.running {
		return nil // Already running
	}

	// Initialize global registry
	if err := core.InitializeGlobalRegistry(m.collector, m.config); err != nil {
		return err
	}

	// Start exporter if available
	if m.exporter != nil {
		if err := m.exporter.Start(); err != nil {
			return err
		}
	}

	m.running = true
	return nil
}

// Stop stops the monitoring system
func (m *Manager) Stop() *types.Error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil // Already stopped
	}

	var lastError *types.Error

	// Stop exporter if available
	if m.exporter != nil {
		if err := m.exporter.Stop(); err != nil {
			lastError = err
		}
	}

	// Shutdown global registry
	if err := core.ShutdownGlobalRegistry(); err != nil {
		lastError = err
	}

	// Shutdown collector
	if m.collector != nil {
		if err := m.collector.Shutdown(); err != nil {
			lastError = err
		}
	}

	m.running = false
	return lastError
}

// RegisterProvider registers a metric provider
func (m *Manager) RegisterProvider(name string, factory types.ProviderFactory, autoInitialize bool, priority int) *types.Error {
	m.mu.RLock()
	initialized := m.initialized
	m.mu.RUnlock()

	if !initialized {
		return types.NewRegistryNotInitializedError()
	}

	return core.RegisterGlobalProvider(name, factory, autoInitialize, priority)
}

// GetCollector returns the collector instance
func (m *Manager) GetCollector() *core.Collector {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.collector
}

// GetRegistry returns the registry instance
func (m *Manager) GetRegistry() *core.Registry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.registry
}

// GetExporter returns the exporter instance
func (m *Manager) GetExporter() *core.Exporter {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.exporter
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *core.MonitoringConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy
	configCopy := *m.config
	return &configCopy
}

// UpdateConfig updates the monitoring configuration
func (m *Manager) UpdateConfig(newConfig *core.MonitoringConfig) *types.Error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if newConfig == nil {
		return types.NewInvalidConfigurationError("config", nil, "configuration cannot be nil")
	}

	// Validate new configuration
	if err := newConfig.Validate(); err != nil {
		return err
	}

	// Store old config for comparison
	oldConfig := m.config

	// Update config
	m.config = newConfig

	// Apply configuration changes
	if m.initialized {
		// Check if we need to restart exporter
		if oldConfig.IsExportEnabled() != newConfig.IsExportEnabled() {
			if m.exporter != nil && !newConfig.IsExportEnabled() {
				// Stop exporter
				if err := m.exporter.Stop(); err != nil {
					return err
				}
				m.exporter = nil
			} else if m.exporter == nil && newConfig.IsExportEnabled() {
				// Start exporter
				exporter, err := core.NewExporter(m.collector, &newConfig.ExportConfig)
				if err != nil {
					return err
				}
				m.exporter = exporter
				if err := m.exporter.Start(); err != nil {
					return err
				}
			}
		}

		// Update collector configuration
		// Note: In a real implementation, you might need to restart the collector
		// or update its internal configuration
	}

	return nil
}

// CollectMetrics triggers a manual metric collection
func (m *Manager) CollectMetrics() *types.Error {
	m.mu.RLock()
	collector := m.collector
	m.mu.RUnlock()

	if collector == nil {
		return types.NewCollectorNotInitializedError()
	}

	return collector.CollectFromProviders()
}

// ExportMetrics exports metrics in the specified format
func (m *Manager) ExportMetrics(format string) (string, *types.Error) {
	m.mu.RLock()
	exporter := m.exporter
	m.mu.RUnlock()

	if exporter == nil {
		return "", types.NewExportFailedError("exporter not initialized")
	}

	switch format {
	case "prometheus":
		return exporter.ExportPrometheus()
	case "json":
		return exporter.ExportJSON()
	default:
		return "", types.NewInvalidConfigurationError("format", format, "unsupported format. Use 'prometheus' or 'json'")
	}
}

// GetStats returns manager statistics
func (m *Manager) GetStats() ManagerStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := m.stats
	stats.UptimeSeconds = time.Now().Unix() - stats.StartTime

	// Update dynamic stats
	if m.collector != nil {
		collectorStats := m.collector.GetStats()
		stats.TotalMetrics = collectorStats.MetricsCollected
		stats.CollectionCycles = collectorStats.BatchOperations
	}

	if m.registry != nil {
		registryStats := m.registry.GetStats()
		stats.ActiveProviders = registryStats.ActiveProviders
	}

	if m.exporter != nil {
		exporterStats := m.exporter.GetStats()
		stats.ExportRequests = exporterStats.RequestsTotal
	}

	return stats
}

// GetProviderHealth returns health status for all providers
func (m *Manager) GetProviderHealth() map[string]types.ProviderHealthStatus {
	m.mu.RLock()
	registry := m.registry
	m.mu.RUnlock()

	if registry == nil {
		return make(map[string]types.ProviderHealthStatus)
	}

	return registry.GetProviderHealth()
}

// GetProviderMetadata returns metadata for all providers
func (m *Manager) GetProviderMetadata() map[string]types.ProviderMetadata {
	m.mu.RLock()
	registry := m.registry
	m.mu.RUnlock()

	if registry == nil {
		return make(map[string]types.ProviderMetadata)
	}

	return registry.GetProviderMetadata()
}

// IsRunning checks if the monitoring system is running
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.running
}

// IsInitialized checks if the monitoring system is initialized
func (m *Manager) IsInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.initialized
}

// Shutdown gracefully shuts down the monitoring system
func (m *Manager) Shutdown() *types.Error {
	return m.Stop()
}

// Factory functions for common use cases

// NewManagerWithConfig creates a new manager with the specified configuration
func NewManagerWithConfig(config *core.MonitoringConfig) (*Manager, *types.Error) {
	return NewManager(config)
}

// NewDefaultManager creates a new manager with default configuration
func NewDefaultManager() (*Manager, *types.Error) {
	config := core.DefaultMonitoringConfig()
	return NewManager(&config)
}

// NewDevelopmentManager creates a new manager with development configuration
func NewDevelopmentManager() (*Manager, *types.Error) {
	config := core.DevelopmentConfig()
	return NewManager(&config)
}

// NewProductionManager creates a new manager with production configuration
func NewProductionManager() (*Manager, *types.Error) {
	config := core.ProductionConfig()
	return NewManager(&config)
}

// NewHighLoadManager creates a new manager with high-load configuration
func NewHighLoadManager() (*Manager, *types.Error) {
	config := core.HighLoadConfig()
	return NewManager(&config)
}

// Global manager instance (optional, for convenience)

var (
	globalManager     *Manager
	globalManagerOnce sync.Once
	globalManagerMu   sync.RWMutex
)

// GetGlobalManager returns the global manager instance
func GetGlobalManager() *Manager {
	globalManagerMu.RLock()
	defer globalManagerMu.RUnlock()

	return globalManager
}

// InitializeGlobalManager initializes the global manager with default configuration
func InitializeGlobalManager() *types.Error {
	return InitializeGlobalManagerWithConfig(nil)
}

// InitializeGlobalManagerWithConfig initializes the global manager with the specified configuration
func InitializeGlobalManagerWithConfig(config *core.MonitoringConfig) *types.Error {
	globalManagerMu.Lock()
	defer globalManagerMu.Unlock()

	if globalManager != nil {
		return nil // Already initialized
	}

	var err *types.Error
	globalManagerOnce.Do(func() {
		if config == nil {
			defaultConfig := core.DefaultMonitoringConfig()
			config = &defaultConfig
		}
		globalManager, err = NewManager(config)
		if err != nil {
			return
		}
		err = globalManager.Start()
	})

	return err
}

// ShutdownGlobalManager shuts down the global manager
func ShutdownGlobalManager() *types.Error {
	globalManagerMu.Lock()
	defer globalManagerMu.Unlock()

	if globalManager == nil {
		return nil // Already shutdown
	}

	err := globalManager.Shutdown()
	globalManager = nil
	return err
}

// RegisterGlobalProvider registers a provider with the global manager
func RegisterGlobalProvider(name string, factory types.ProviderFactory, autoInitialize bool, priority int) *types.Error {
	manager := GetGlobalManager()
	if manager == nil {
		return types.NewRegistryNotInitializedError()
	}

	return manager.RegisterProvider(name, factory, autoInitialize, priority)
}
