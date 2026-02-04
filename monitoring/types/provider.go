package types

import (
	"time"

	cd "github.com/muidea/magicCommon/def"
)

// MetricProvider is the interface that must be implemented by modules
// that want to provide metrics to the monitoring system.
type MetricProvider interface {
	// Name returns the unique name of the metric provider
	Name() string

	// Metrics returns the list of metric definitions provided by this provider
	Metrics() []MetricDefinition

	// Init is called when the provider is registered with the monitoring system
	// It receives a collector instance that the provider can use to collect metrics
	Init(collector any) *cd.Error

	// Collect is called periodically to collect current metric values
	// Providers should update their metrics and return them
	Collect() ([]Metric, *cd.Error)

	// Shutdown is called when the provider is being unregistered
	// Providers should clean up any resources they're using
	Shutdown() *cd.Error

	// GetMetadata returns additional metadata about the provider
	GetMetadata() ProviderMetadata
}

// ProviderMetadata contains metadata about a metric provider
type ProviderMetadata struct {
	// Version of the provider
	Version string `json:"version"`

	// Description of what metrics this provider collects
	Description string `json:"description"`

	// Tags for categorizing providers
	Tags []string `json:"tags,omitempty"`

	// Dependencies on other providers (if any)
	Dependencies []string `json:"dependencies,omitempty"`

	// Configuration schema for provider-specific configuration
	ConfigSchema interface{} `json:"config_schema,omitempty"`

	// Health status of the provider
	HealthStatus ProviderHealthStatus `json:"health_status"`

	// Last collection timestamp
	LastCollectionTime time.Time `json:"last_collection_time"`

	// Collection statistics
	CollectionStats ProviderCollectionStats `json:"collection_stats"`
}

// ProviderHealthStatus represents the health status of a provider
type ProviderHealthStatus string

const (
	// ProviderHealthy indicates the provider is functioning normally
	ProviderHealthy ProviderHealthStatus = "healthy"
	// ProviderDegraded indicates the provider is functioning but with issues
	ProviderDegraded ProviderHealthStatus = "degraded"
	// ProviderUnhealthy indicates the provider is not functioning
	ProviderUnhealthy ProviderHealthStatus = "unhealthy"
	// ProviderUnknown indicates the provider health status is unknown
	ProviderUnknown ProviderHealthStatus = "unknown"
)

// ProviderCollectionStats contains collection statistics for a provider
type ProviderCollectionStats struct {
	// Total number of collections performed
	TotalCollections int64 `json:"total_collections"`

	// Number of successful collections
	SuccessfulCollections int64 `json:"successful_collections"`

	// Number of failed collections
	FailedCollections int64 `json:"failed_collections"`

	// Total number of metrics collected
	TotalMetricsCollected int64 `json:"total_metrics_collected"`

	// Average collection duration in milliseconds
	AverageCollectionDuration float64 `json:"average_collection_duration"`

	// Last collection duration in milliseconds
	LastCollectionDuration float64 `json:"last_collection_duration"`

	// Last error message (if any)
	LastErrorMessage string `json:"last_error_message,omitempty"`

	// Last error time (if any)
	LastErrorTime time.Time `json:"last_error_time,omitempty"`
}

// BaseProvider provides a base implementation of MetricProvider
// that other providers can embed to get default implementations
type BaseProvider struct {
	name     string
	metadata ProviderMetadata
}

// NewBaseProvider creates a new base provider
func NewBaseProvider(name, version, description string) *BaseProvider {
	return &BaseProvider{
		name: name,
		metadata: ProviderMetadata{
			Version:      version,
			Description:  description,
			HealthStatus: ProviderUnknown,
			CollectionStats: ProviderCollectionStats{
				TotalCollections:          0,
				SuccessfulCollections:     0,
				FailedCollections:         0,
				TotalMetricsCollected:     0,
				AverageCollectionDuration: 0,
				LastCollectionDuration:    0,
			},
		},
	}
}

// Name returns the provider name
func (p *BaseProvider) Name() string {
	return p.name
}

// Init provides a default implementation that does nothing
func (p *BaseProvider) Init(collector interface{}) *cd.Error {
	// Default implementation does nothing
	return nil
}

// Collect provides a default implementation that returns no metrics
func (p *BaseProvider) Collect() ([]Metric, *cd.Error) {
	// Default implementation returns no metrics
	return []Metric{}, nil
}

// Shutdown provides a default implementation that does nothing
func (p *BaseProvider) Shutdown() *cd.Error {
	// Default implementation does nothing
	return nil
}

// GetMetadata returns the provider metadata
func (p *BaseProvider) GetMetadata() ProviderMetadata {
	return p.metadata
}

// UpdateHealthStatus updates the provider's health status
func (p *BaseProvider) UpdateHealthStatus(status ProviderHealthStatus) {
	p.metadata.HealthStatus = status
}

// UpdateCollectionStats updates the provider's collection statistics
func (p *BaseProvider) UpdateCollectionStats(success bool, duration time.Duration, metricsCollected int) {
	p.metadata.LastCollectionTime = time.Now()
	p.metadata.CollectionStats.TotalCollections++

	if success {
		p.metadata.CollectionStats.SuccessfulCollections++
		p.metadata.CollectionStats.LastErrorMessage = ""
		p.metadata.CollectionStats.LastErrorTime = time.Time{}
	} else {
		p.metadata.CollectionStats.FailedCollections++
	}

	p.metadata.CollectionStats.TotalMetricsCollected += int64(metricsCollected)

	// Update average collection duration
	oldTotal := float64(p.metadata.CollectionStats.TotalCollections-1) * p.metadata.CollectionStats.AverageCollectionDuration
	p.metadata.CollectionStats.AverageCollectionDuration = (oldTotal + float64(duration.Milliseconds())) / float64(p.metadata.CollectionStats.TotalCollections)

	p.metadata.CollectionStats.LastCollectionDuration = float64(duration.Milliseconds())
}

// UpdateLastError updates the last error information
func (p *BaseProvider) UpdateLastError(err *cd.Error) {
	if err != nil {
		p.metadata.CollectionStats.LastErrorMessage = err.Error()
		p.metadata.CollectionStats.LastErrorTime = time.Now()
		p.UpdateHealthStatus(ProviderDegraded)
	}
}

// AddTag adds a tag to the provider
func (p *BaseProvider) AddTag(tag string) {
	for _, existingTag := range p.metadata.Tags {
		if existingTag == tag {
			return // Tag already exists
		}
	}
	p.metadata.Tags = append(p.metadata.Tags, tag)
}

// AddDependency adds a dependency to the provider
func (p *BaseProvider) AddDependency(dependency string) {
	for _, existingDep := range p.metadata.Dependencies {
		if existingDep == dependency {
			return // Dependency already exists
		}
	}
	p.metadata.Dependencies = append(p.metadata.Dependencies, dependency)
}

// SetConfigSchema sets the configuration schema for the provider
func (p *BaseProvider) SetConfigSchema(schema interface{}) {
	p.metadata.ConfigSchema = schema
}

// ProviderFactory is a function that creates a new MetricProvider
type ProviderFactory func() MetricProvider

// ProviderRegistryEntry represents an entry in the provider registry
type ProviderRegistryEntry struct {
	// Factory function to create the provider
	Factory ProviderFactory

	// Whether the provider should be auto-initialized
	AutoInitialize bool

	// Priority for initialization (lower numbers initialized first)
	Priority int

	// Whether the provider is enabled
	Enabled bool
}

// NewProviderRegistryEntry creates a new provider registry entry
func NewProviderRegistryEntry(factory ProviderFactory, autoInitialize bool, priority int) ProviderRegistryEntry {
	return ProviderRegistryEntry{
		Factory:        factory,
		AutoInitialize: autoInitialize,
		Priority:       priority,
		Enabled:        true,
	}
}
