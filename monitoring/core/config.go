package core

import (
	"time"

	"github.com/muidea/magicCommon/monitoring/types"
)

// DetailLevel defines the level of monitoring detail
type DetailLevel string

const (
	// DetailLevelBasic collects only essential metrics
	DetailLevelBasic DetailLevel = "basic"
	// DetailLevelStandard collects standard operational metrics
	DetailLevelStandard DetailLevel = "standard"
	// DetailLevelDetailed collects comprehensive metrics including performance breakdowns
	DetailLevelDetailed DetailLevel = "detailed"
)

// MonitoringConfig holds configuration for the unified monitoring system
type MonitoringConfig struct {
	// Enabled controls whether monitoring is active
	Enabled bool `json:"enabled"`

	// Namespace for metric names (e.g., "magicorm", "appname")
	Namespace string `json:"namespace"`

	// SamplingRate controls the rate of metric collection (0.0-1.0)
	SamplingRate float64 `json:"sampling_rate"`

	// AsyncCollection enables asynchronous metric collection
	AsyncCollection bool `json:"async_collection"`
	// CollectionInterval controls how often metrics are collected asynchronously
	CollectionInterval time.Duration `json:"collection_interval"`

	// RetentionPeriod controls how long metrics are retained
	RetentionPeriod time.Duration `json:"retention_period"`

	// Detail level for metrics collection
	DetailLevel DetailLevel `json:"detail_level"`

	// Export configuration
	ExportConfig ExportConfig `json:"export_config"`

	// Performance optimization
	BatchSize          int           `json:"batch_size"`
	BufferSize         int           `json:"buffer_size"`
	MaxConcurrentTasks int           `json:"max_concurrent_tasks"`
	Timeout            time.Duration `json:"timeout"`

	// Provider-specific configuration
	ProviderConfigs map[string]any `json:"provider_configs,omitempty"`

	// Environment-specific settings
	Environment string `json:"environment"`
}

// ExportConfig holds configuration for metric export
type ExportConfig struct {
	// Enabled controls whether metrics are exported
	Enabled bool `json:"enabled"`

	// HTTP server configuration
	Port            int    `json:"port"`
	Path            string `json:"path"`
	HealthCheckPath string `json:"health_check_path"`
	MetricsPath     string `json:"metrics_path"`
	InfoPath        string `json:"info_path"`

	// Format support
	EnablePrometheus bool `json:"enable_prometheus"`
	EnableJSON       bool `json:"enable_json"`

	// Export intervals
	RefreshInterval time.Duration `json:"refresh_interval"`
	ScrapeTimeout   time.Duration `json:"scrape_timeout"`

	// Security
	EnableTLS    bool     `json:"enable_tls"`
	TLSCertPath  string   `json:"tls_cert_path"`
	TLSKeyPath   string   `json:"tls_key_path"`
	EnableAuth   bool     `json:"enable_auth"`
	AuthToken    string   `json:"auth_token"`
	AllowedHosts []string `json:"allowed_hosts"`
}

// DefaultMonitoringConfig returns the default monitoring configuration
func DefaultMonitoringConfig() MonitoringConfig {
	return MonitoringConfig{
		Enabled:            true,
		Namespace:          "app",
		SamplingRate:       1.0,
		AsyncCollection:    true,
		CollectionInterval: 30 * time.Second,
		RetentionPeriod:    24 * time.Hour,
		DetailLevel:        DetailLevelStandard,
		ExportConfig:       DefaultExportConfig(),
		BatchSize:          100,
		BufferSize:         1000,
		MaxConcurrentTasks: 10,
		Timeout:            10 * time.Second,
		ProviderConfigs:    make(map[string]any),
		Environment:        "development",
	}
}

// DefaultExportConfig returns the default export configuration
func DefaultExportConfig() ExportConfig {
	return ExportConfig{
		Enabled:          true,
		Port:             9090,
		Path:             "/metrics",
		HealthCheckPath:  "/health",
		MetricsPath:      "/metrics/json",
		InfoPath:         "/",
		EnablePrometheus: true,
		EnableJSON:       true,
		RefreshInterval:  30 * time.Second,
		ScrapeTimeout:    10 * time.Second,
		EnableTLS:        false,
		EnableAuth:       false,
		AllowedHosts:     []string{},
	}
}

// DevelopmentConfig returns configuration suitable for development
func DevelopmentConfig() MonitoringConfig {
	config := DefaultMonitoringConfig()
	config.SamplingRate = 0.1 // 10% sampling in development
	config.DetailLevel = DetailLevelBasic
	config.ExportConfig.Enabled = false
	config.AsyncCollection = false // Synchronous for easier debugging
	config.Environment = "development"
	return config
}

// ProductionConfig returns configuration suitable for production
func ProductionConfig() MonitoringConfig {
	config := DefaultMonitoringConfig()
	config.SamplingRate = 0.5 // 50% sampling in production
	config.DetailLevel = DetailLevelStandard
	config.ExportConfig.Enabled = true
	config.ExportConfig.EnableAuth = true
	config.ExportConfig.EnableTLS = true
	config.BatchSize = 500
	config.BufferSize = 5000
	config.MaxConcurrentTasks = 50
	config.Environment = "production"
	return config
}

// HighLoadConfig returns configuration for high-load environments
func HighLoadConfig() MonitoringConfig {
	config := ProductionConfig()
	config.SamplingRate = 0.1 // 10% sampling under high load
	config.DetailLevel = DetailLevelBasic
	config.ExportConfig.RefreshInterval = 60 * time.Second
	config.BatchSize = 1000
	config.BufferSize = 10000
	config.MaxConcurrentTasks = 100
	config.Environment = "highload"
	return config
}

// Validate validates the monitoring configuration
func (c *MonitoringConfig) Validate() *types.Error {
	if c.SamplingRate < 0 || c.SamplingRate > 1 {
		return types.NewInvalidConfigurationError("sampling_rate", c.SamplingRate, "must be between 0 and 1")
	}

	if c.CollectionInterval <= 0 {
		return types.NewInvalidConfigurationError("collection_interval", c.CollectionInterval, "must be positive")
	}

	if c.RetentionPeriod <= 0 {
		return types.NewInvalidConfigurationError("retention_period", c.RetentionPeriod, "must be positive")
	}

	if c.BatchSize <= 0 {
		return types.NewInvalidConfigurationError("batch_size", c.BatchSize, "must be positive")
	}

	if c.BufferSize <= 0 {
		return types.NewInvalidConfigurationError("buffer_size", c.BufferSize, "must be positive")
	}

	if c.MaxConcurrentTasks <= 0 {
		return types.NewInvalidConfigurationError("max_concurrent_tasks", c.MaxConcurrentTasks, "must be positive")
	}

	if c.Timeout <= 0 {
		return types.NewInvalidConfigurationError("timeout", c.Timeout, "must be positive")
	}

	if c.Namespace == "" {
		return types.NewInvalidConfigurationError("namespace", c.Namespace, "cannot be empty")
	}

	// Validate detail level
	switch c.DetailLevel {
	case DetailLevelBasic, DetailLevelStandard, DetailLevelDetailed:
		// Valid
	default:
		return types.NewInvalidConfigurationError("detail_level", c.DetailLevel, "must be one of: basic, standard, detailed")
	}

	return c.ExportConfig.Validate()
}

// Validate validates the export configuration
func (c *ExportConfig) Validate() *types.Error {
	if c.Port <= 0 || c.Port > 65535 {
		return types.NewInvalidConfigurationError("port", c.Port, "must be between 1 and 65535")
	}

	if c.Path == "" {
		return types.NewInvalidConfigurationError("path", c.Path, "cannot be empty")
	}

	if c.RefreshInterval <= 0 {
		return types.NewInvalidConfigurationError("refresh_interval", c.RefreshInterval, "must be positive")
	}

	if c.ScrapeTimeout <= 0 {
		return types.NewInvalidConfigurationError("scrape_timeout", c.ScrapeTimeout, "must be positive")
	}

	if c.EnableTLS {
		if c.TLSCertPath == "" {
			return types.NewInvalidConfigurationError("tls_cert_path", c.TLSCertPath, "required when TLS is enabled")
		}
		if c.TLSKeyPath == "" {
			return types.NewInvalidConfigurationError("tls_key_path", c.TLSKeyPath, "required when TLS is enabled")
		}
	}

	if c.EnableAuth && c.AuthToken == "" {
		return types.NewInvalidConfigurationError("auth_token", c.AuthToken, "required when auth is enabled")
	}

	return nil
}

// ShouldSample determines if a metric should be collected based on sampling rate
func (c *MonitoringConfig) ShouldSample() bool {
	if !c.Enabled {
		return false
	}
	if c.SamplingRate >= 1.0 {
		return true
	}
	if c.SamplingRate <= 0 {
		return false
	}
	// Simple deterministic sampling based on metric name hash
	// In production, you might want to use a proper sampling algorithm
	return true // Placeholder - will be implemented with proper sampling
}

// GetProviderConfig retrieves provider-specific configuration
func (c *MonitoringConfig) GetProviderConfig(providerName string) interface{} {
	if config, exists := c.ProviderConfigs[providerName]; exists {
		return config
	}
	return nil
}

// SetProviderConfig sets provider-specific configuration
func (c *MonitoringConfig) SetProviderConfig(providerName string, config interface{}) {
	if c.ProviderConfigs == nil {
		c.ProviderConfigs = make(map[string]interface{})
	}
	c.ProviderConfigs[providerName] = config
}

// IsExportEnabled checks if metric export is enabled
func (c *MonitoringConfig) IsExportEnabled() bool {
	return c.Enabled && c.ExportConfig.Enabled
}

// GetEnvironmentConfig returns configuration for the current environment
func GetEnvironmentConfig(environment string) MonitoringConfig {
	switch environment {
	case "development":
		return DevelopmentConfig()
	case "production":
		return ProductionConfig()
	case "highload":
		return HighLoadConfig()
	default:
		return DefaultMonitoringConfig()
	}
}

// MergeConfigs merges multiple configurations, with later configs overriding earlier ones
func MergeConfigs(configs ...MonitoringConfig) MonitoringConfig {
	if len(configs) == 0 {
		return DefaultMonitoringConfig()
	}

	result := configs[0]
	for i := 1; i < len(configs); i++ {
		config := configs[i]

		// Merge simple fields
		if config.Enabled != result.Enabled {
			result.Enabled = config.Enabled
		}
		if config.Namespace != "" {
			result.Namespace = config.Namespace
		}
		if config.SamplingRate != result.SamplingRate {
			result.SamplingRate = config.SamplingRate
		}
		if config.AsyncCollection != result.AsyncCollection {
			result.AsyncCollection = config.AsyncCollection
		}
		if config.CollectionInterval != result.CollectionInterval {
			result.CollectionInterval = config.CollectionInterval
		}
		if config.RetentionPeriod != result.RetentionPeriod {
			result.RetentionPeriod = config.RetentionPeriod
		}
		if config.DetailLevel != "" {
			result.DetailLevel = config.DetailLevel
		}
		if config.BatchSize != result.BatchSize {
			result.BatchSize = config.BatchSize
		}
		if config.BufferSize != result.BufferSize {
			result.BufferSize = config.BufferSize
		}
		if config.MaxConcurrentTasks != result.MaxConcurrentTasks {
			result.MaxConcurrentTasks = config.MaxConcurrentTasks
		}
		if config.Timeout != result.Timeout {
			result.Timeout = config.Timeout
		}
		if config.Environment != "" {
			result.Environment = config.Environment
		}

		// Merge provider configs
		for name, providerConfig := range config.ProviderConfigs {
			result.SetProviderConfig(name, providerConfig)
		}

		// Merge export config
		result.ExportConfig = MergeExportConfigs(result.ExportConfig, config.ExportConfig)
	}

	return result
}

// MergeExportConfigs merges export configurations
func MergeExportConfigs(configs ...ExportConfig) ExportConfig {
	if len(configs) == 0 {
		return DefaultExportConfig()
	}

	result := configs[0]
	for i := 1; i < len(configs); i++ {
		config := configs[i]

		if config.Enabled != result.Enabled {
			result.Enabled = config.Enabled
		}
		if config.Port != 0 {
			result.Port = config.Port
		}
		if config.Path != "" {
			result.Path = config.Path
		}
		if config.HealthCheckPath != "" {
			result.HealthCheckPath = config.HealthCheckPath
		}
		if config.MetricsPath != "" {
			result.MetricsPath = config.MetricsPath
		}
		if config.InfoPath != "" {
			result.InfoPath = config.InfoPath
		}
		if config.EnablePrometheus != result.EnablePrometheus {
			result.EnablePrometheus = config.EnablePrometheus
		}
		if config.EnableJSON != result.EnableJSON {
			result.EnableJSON = config.EnableJSON
		}
		if config.RefreshInterval != 0 {
			result.RefreshInterval = config.RefreshInterval
		}
		if config.ScrapeTimeout != 0 {
			result.ScrapeTimeout = config.ScrapeTimeout
		}
		if config.EnableTLS != result.EnableTLS {
			result.EnableTLS = config.EnableTLS
		}
		if config.TLSCertPath != "" {
			result.TLSCertPath = config.TLSCertPath
		}
		if config.TLSKeyPath != "" {
			result.TLSKeyPath = config.TLSKeyPath
		}
		if config.EnableAuth != result.EnableAuth {
			result.EnableAuth = config.EnableAuth
		}
		if config.AuthToken != "" {
			result.AuthToken = config.AuthToken
		}
		if len(config.AllowedHosts) > 0 {
			result.AllowedHosts = config.AllowedHosts
		}
	}

	return result
}
