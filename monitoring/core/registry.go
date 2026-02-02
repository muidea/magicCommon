package core

import (
	"fmt"
	"sort"
	"sync"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/monitoring/types"
)

// Registry manages metric providers and their registration
type Registry struct {
	mu sync.RWMutex

	// Provider registry
	providers map[string]types.ProviderRegistryEntry

	// Active providers (initialized and running)
	activeProviders map[string]types.MetricProvider

	// Collector reference
	collector *Collector

	// Configuration
	config *MonitoringConfig

	// Registry statistics
	stats RegistryStats
}

// RegistryStats holds registry statistics
type RegistryStats struct {
	TotalProvidersRegistered int64 `json:"total_providers_registered"`
	ActiveProviders          int64 `json:"active_providers"`
	FailedRegistrations      int64 `json:"failed_registrations"`
	SuccessfulRegistrations  int64 `json:"successful_registrations"`
}

// NewRegistry creates a new metric registry
func NewRegistry(collector *Collector, config *MonitoringConfig) (*Registry, *types.Error) {
	if collector == nil {
		return nil, types.NewCollectorNotInitializedError()
	}

	if config == nil {
		defaultConfig := DefaultMonitoringConfig()
		config = &defaultConfig
	}

	registry := &Registry{
		providers:       make(map[string]types.ProviderRegistryEntry),
		activeProviders: make(map[string]types.MetricProvider),
		collector:       collector,
		config:          config,
		stats:           RegistryStats{},
	}

	return registry, nil
}

// Register registers a provider factory with the registry
func (r *Registry) Register(name string, entry types.ProviderRegistryEntry) *types.Error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; exists {
		return types.NewProviderAlreadyRegisteredError(name)
	}

	r.providers[name] = entry
	r.stats.TotalProvidersRegistered++

	// Auto-initialize if configured
	if entry.AutoInitialize && entry.Enabled {
		if err := r.initializeProvider(name, entry); err != nil {
			r.stats.FailedRegistrations++
			return err
		}
		r.stats.SuccessfulRegistrations++
	}

	return nil
}

// Unregister removes a provider from the registry
func (r *Registry) Unregister(name string) *types.Error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if provider exists
	if _, exists := r.providers[name]; !exists {
		return types.NewProviderNotFoundError(name)
	}

	// Shutdown if active
	if provider, isActive := r.activeProviders[name]; isActive {
		if err := provider.Shutdown(); err != nil {
			return err
		}
		delete(r.activeProviders, name)
		r.stats.ActiveProviders--
	}

	// Remove from registry
	delete(r.providers, name)
	r.stats.TotalProvidersRegistered--

	return nil
}

// InitializeProvider initializes a registered provider
func (r *Registry) InitializeProvider(name string) *types.Error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.providers[name]
	if !exists {
		return types.NewProviderNotFoundError(name)
	}

	if !entry.Enabled {
		return types.NewError(cd.InvalidOperation, "provider '"+name+"' is disabled")
	}

	// Check if already active
	if _, isActive := r.activeProviders[name]; isActive {
		return nil // Already initialized
	}

	return r.initializeProvider(name, entry)
}

// InitializeAll initializes all registered providers
func (r *Registry) InitializeAll() *types.Error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Sort providers by priority
	providers := r.getSortedProviders()

	var lastError *types.Error
	for _, providerInfo := range providers {
		if !providerInfo.entry.Enabled {
			continue
		}

		if _, isActive := r.activeProviders[providerInfo.name]; isActive {
			continue // Already initialized
		}

		if err := r.initializeProvider(providerInfo.name, providerInfo.entry); err != nil {
			lastError = err
			r.stats.FailedRegistrations++
		} else {
			r.stats.SuccessfulRegistrations++
		}
	}

	return lastError
}

// ShutdownProvider shuts down an active provider
func (r *Registry) ShutdownProvider(name string) *types.Error {
	r.mu.Lock()
	defer r.mu.Unlock()

	provider, exists := r.activeProviders[name]
	if !exists {
		return types.NewProviderNotFoundError(name)
	}

	if err := provider.Shutdown(); err != nil {
		return err
	}

	delete(r.activeProviders, name)
	r.stats.ActiveProviders--

	return nil
}

// ShutdownAll shuts down all active providers
func (r *Registry) ShutdownAll() *types.Error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var lastError *types.Error
	for name, provider := range r.activeProviders {
		if err := provider.Shutdown(); err != nil {
			lastError = err
		}
		delete(r.activeProviders, name)
	}

	r.stats.ActiveProviders = 0
	return lastError
}

// EnableProvider enables a disabled provider
func (r *Registry) EnableProvider(name string) *types.Error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.providers[name]
	if !exists {
		return types.NewProviderNotFoundError(name)
	}

	entry.Enabled = true
	r.providers[name] = entry

	// Auto-initialize if configured
	if entry.AutoInitialize {
		if err := r.initializeProvider(name, entry); err != nil {
			return err
		}
	}

	return nil
}

// DisableProvider disables an enabled provider
func (r *Registry) DisableProvider(name string) *types.Error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.providers[name]
	if !exists {
		return types.NewProviderNotFoundError(name)
	}

	entry.Enabled = false
	r.providers[name] = entry

	// Shutdown if active
	if provider, isActive := r.activeProviders[name]; isActive {
		if err := provider.Shutdown(); err != nil {
			return err
		}
		delete(r.activeProviders, name)
		r.stats.ActiveProviders--
	}

	return nil
}

// GetProvider returns a provider by name
func (r *Registry) GetProvider(name string) (types.MetricProvider, *types.Error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.activeProviders[name]
	if !exists {
		return nil, types.NewProviderNotFoundError(name)
	}
	return provider, nil
}

// GetProviders returns all active providers
func (r *Registry) GetProviders() map[string]types.MetricProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy
	result := make(map[string]types.MetricProvider)
	for name, provider := range r.activeProviders {
		result[name] = provider
	}
	return result
}

// GetProviderMetadata returns metadata for all providers
func (r *Registry) GetProviderMetadata() map[string]types.ProviderMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]types.ProviderMetadata)
	for name, provider := range r.activeProviders {
		result[name] = provider.GetMetadata()
	}
	return result
}

// GetRegistryEntries returns all registry entries
func (r *Registry) GetRegistryEntries() map[string]types.ProviderRegistryEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy
	result := make(map[string]types.ProviderRegistryEntry)
	for name, entry := range r.providers {
		result[name] = entry
	}
	return result
}

// GetStats returns registry statistics
func (r *Registry) GetStats() RegistryStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := r.stats
	stats.ActiveProviders = int64(len(r.activeProviders))
	return stats
}

// CollectFromAll collects metrics from all active providers
func (r *Registry) CollectFromAll() *types.Error {
	r.mu.RLock()
	providers := make([]types.MetricProvider, 0, len(r.activeProviders))
	for _, provider := range r.activeProviders {
		providers = append(providers, provider)
	}
	r.mu.RUnlock()

	var lastError *types.Error
	for _, provider := range providers {
		if _, err := provider.Collect(); err != nil {
			lastError = err
		}
	}

	return lastError
}

// ValidateDependencies validates provider dependencies
func (r *Registry) ValidateDependencies() *types.Error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, provider := range r.activeProviders {
		metadata := provider.GetMetadata()
		for _, dep := range metadata.Dependencies {
			if _, exists := r.activeProviders[dep]; !exists {
				return types.NewError(cd.InvalidOperation,
					"provider '"+name+"' depends on missing provider '"+dep+"'")
			}
		}
	}

	return nil
}

// GetProviderHealth returns health status for all providers
func (r *Registry) GetProviderHealth() map[string]types.ProviderHealthStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]types.ProviderHealthStatus)
	for name, provider := range r.activeProviders {
		result[name] = provider.GetMetadata().HealthStatus
	}
	return result
}

// Private helper methods

type providerInfo struct {
	name     string
	entry    types.ProviderRegistryEntry
	priority int
}

func (r *Registry) getSortedProviders() []providerInfo {
	providers := make([]providerInfo, 0, len(r.providers))
	for name, entry := range r.providers {
		providers = append(providers, providerInfo{
			name:     name,
			entry:    entry,
			priority: entry.Priority,
		})
	}

	// Sort by priority (lower numbers first)
	sort.Slice(providers, func(i, j int) bool {
		if providers[i].priority == providers[j].priority {
			return providers[i].name < providers[j].name
		}
		return providers[i].priority < providers[j].priority
	})

	return providers
}

func (r *Registry) initializeProvider(name string, entry types.ProviderRegistryEntry) *types.Error {
	// Create provider instance
	provider := entry.Factory()

	// Check dependencies before initialization
	if err := r.checkDependencies(provider); err != nil {
		return err
	}

	// Register provider with collector (collector will call Init)
	if err := r.collector.RegisterProvider(provider); err != nil {
		return err
	}

	// Add to active providers
	r.activeProviders[name] = provider
	r.stats.ActiveProviders++

	return nil
}

// checkDependencies verifies that all provider dependencies are satisfied
func (r *Registry) checkDependencies(provider types.MetricProvider) *types.Error {
	metadata := provider.GetMetadata()

	for _, dep := range metadata.Dependencies {
		// Check if dependency is registered
		if _, exists := r.providers[dep]; !exists {
			return types.NewError(cd.InvalidOperation,
				fmt.Sprintf("provider '%s' requires dependency '%s' which is not registered",
					provider.Name(), dep))
		}

		// Check if dependency is enabled
		if entry, exists := r.providers[dep]; exists && !entry.Enabled {
			return types.NewError(cd.InvalidOperation,
				fmt.Sprintf("provider '%s' requires dependency '%s' which is disabled",
					provider.Name(), dep))
		}

		// Check if dependency is initialized (if it should be)
		if entry, exists := r.providers[dep]; exists && entry.AutoInitialize {
			if _, isActive := r.activeProviders[dep]; !isActive {
				return types.NewError(cd.InvalidOperation,
					fmt.Sprintf("provider '%s' requires dependency '%s' which is not initialized",
						provider.Name(), dep))
			}
		}
	}

	return nil
}

// Global registry functions

var (
	globalRegistry     *Registry
	globalRegistryOnce sync.Once
	globalRegistryMu   sync.RWMutex
)

// GetGlobalRegistry returns the global registry instance
func GetGlobalRegistry(collector *Collector, config *MonitoringConfig) (*Registry, *types.Error) {
	globalRegistryMu.Lock()
	defer globalRegistryMu.Unlock()

	if globalRegistry == nil {
		if collector == nil || config == nil {
			return nil, types.NewRegistryNotInitializedError()
		}

		var err *types.Error
		globalRegistryOnce.Do(func() {
			globalRegistry, err = NewRegistry(collector, config)
		})

		if err != nil {
			return nil, err
		}
	}

	return globalRegistry, nil
}

// RegisterGlobalProvider registers a provider with the global registry
func RegisterGlobalProvider(name string, factory types.ProviderFactory, autoInitialize bool, priority int) *types.Error {
	globalRegistryMu.RLock()
	registry := globalRegistry
	globalRegistryMu.RUnlock()

	if registry == nil {
		return types.NewRegistryNotInitializedError()
	}

	entry := types.NewProviderRegistryEntry(factory, autoInitialize, priority)
	return registry.Register(name, entry)
}

// InitializeGlobalRegistry initializes the global registry with default providers
func InitializeGlobalRegistry(collector *Collector, config *MonitoringConfig) *types.Error {
	globalRegistryMu.Lock()
	defer globalRegistryMu.Unlock()

	if globalRegistry != nil {
		return nil // Already initialized
	}

	var err *types.Error
	globalRegistry, err = NewRegistry(collector, config)
	if err != nil {
		return err
	}

	// Initialize all auto-initialize providers
	return globalRegistry.InitializeAll()
}

// ShutdownGlobalRegistry shuts down the global registry
func ShutdownGlobalRegistry() *types.Error {
	globalRegistryMu.Lock()
	defer globalRegistryMu.Unlock()

	if globalRegistry == nil {
		return nil // Already shutdown
	}

	err := globalRegistry.ShutdownAll()
	globalRegistry = nil
	return err
}
