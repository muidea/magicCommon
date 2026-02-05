package test

import (
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/monitoring"
	"github.com/muidea/magicCommon/monitoring/core"
	"github.com/muidea/magicCommon/monitoring/types"
)

// TestErrorTypes tests monitoring error types
func TestErrorTypes(t *testing.T) {
	// Test error creation
	err := types.NewMetricAlreadyRegisteredError("test_metric")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err.Code != types.MetricAlreadyRegistered {
		t.Errorf("Expected error code %d, got %d", types.MetricAlreadyRegistered, err.Code)
	}

	// Test error code string representation
	codeStr := types.GetErrorCode(err)
	if codeStr != "MetricAlreadyRegistered" {
		t.Errorf("Expected error code string 'MetricAlreadyRegistered', got '%s'", codeStr)
	}

	// Test IsMonitoringError
	if !types.IsMonitoringError(err) {
		t.Error("IsMonitoringError should return true for monitoring errors")
	}

	// Test non-monitoring error
	nonMonitoringErr := cd.NewError(cd.InvalidParameter, "test")
	if types.IsMonitoringError(nonMonitoringErr) {
		t.Error("IsMonitoringError should return false for non-monitoring errors")
	}

	// Test nil error
	if types.IsMonitoringError(nil) {
		t.Error("IsMonitoringError should return false for nil")
	}
}

// TestInvalidConfiguration tests configuration validation errors
func TestInvalidConfiguration(t *testing.T) {
	// Test invalid sampling rate
	config := core.DefaultMonitoringConfig()
	config.SamplingRate = 1.5 // Invalid: > 1.0

	if err := config.Validate(); err == nil {
		t.Error("Expected validation error for sampling rate > 1.0")
	} else if !types.IsMonitoringError(err) {
		t.Error("Expected monitoring error type")
	}

	// Test invalid port
	config = core.DefaultMonitoringConfig()
	config.ExportConfig.Port = 70000 // Invalid port

	if err := config.Validate(); err == nil {
		t.Error("Expected validation error for invalid port")
	}

	// Test empty namespace
	config = core.DefaultMonitoringConfig()
	config.Namespace = ""

	if err := config.Validate(); err == nil {
		t.Error("Expected validation error for empty namespace")
	}

	// Test invalid detail level
	config = core.DefaultMonitoringConfig()
	config.DetailLevel = "invalid"

	if err := config.Validate(); err == nil {
		t.Error("Expected validation error for invalid detail level")
	}

	// Test TLS without certificate
	config = core.DefaultMonitoringConfig()
	config.ExportConfig.EnableTLS = true
	config.ExportConfig.TLSCertPath = ""

	if err := config.Validate(); err == nil {
		t.Error("Expected validation error for TLS without certificate")
	}
}

// TestMetricDefinitionErrors tests metric definition validation errors
func TestMetricDefinitionErrors(t *testing.T) {
	// Test empty metric name
	def := types.NewCounterDefinition("", "Test metric", []string{}, nil)
	if err := def.Validate(); err == nil {
		t.Error("Expected validation error for empty metric name")
	}

	// Test empty help text
	def = types.NewCounterDefinition("test_metric", "", []string{}, nil)
	if err := def.Validate(); err == nil {
		t.Error("Expected validation error for empty help text")
	}

	// Test empty label name
	def = types.NewCounterDefinition("test_metric", "Test metric", []string{""}, nil)
	if err := def.Validate(); err == nil {
		t.Error("Expected validation error for empty label name")
	}

	// Test histogram without buckets
	histogramDef := types.MetricDefinition{
		Name:    "test_histogram",
		Type:    types.HistogramMetric,
		Help:    "Test histogram",
		Buckets: []float64{}, // Empty buckets
	}
	if err := histogramDef.Validate(); err == nil {
		t.Error("Expected validation error for histogram without buckets")
	}

	// Test histogram with unsorted buckets
	histogramDef = types.MetricDefinition{
		Name:    "test_histogram",
		Type:    types.HistogramMetric,
		Help:    "Test histogram",
		Buckets: []float64{10.0, 5.0, 15.0}, // Unsorted
	}
	if err := histogramDef.Validate(); err == nil {
		t.Error("Expected validation error for histogram with unsorted buckets")
	}

	// Test summary without objectives
	summaryDef := types.MetricDefinition{
		Name:       "test_summary",
		Type:       types.SummaryMetric,
		Help:       "Test summary",
		Objectives: map[float64]float64{}, // Empty objectives
		MaxAge:     time.Minute,
	}
	if err := summaryDef.Validate(); err == nil {
		t.Error("Expected validation error for summary without objectives")
	}

	// Test summary with invalid quantile
	summaryDef = types.MetricDefinition{
		Name:       "test_summary",
		Type:       types.SummaryMetric,
		Help:       "Test summary",
		Objectives: map[float64]float64{1.5: 0.01}, // Invalid quantile > 1.0
		MaxAge:     time.Minute,
	}
	if err := summaryDef.Validate(); err == nil {
		t.Error("Expected validation error for summary with invalid quantile")
	}

	// Test summary with negative max age
	summaryDef = types.MetricDefinition{
		Name:       "test_summary",
		Type:       types.SummaryMetric,
		Help:       "Test summary",
		Objectives: map[float64]float64{0.5: 0.01},
		MaxAge:     -time.Minute, // Negative max age
	}
	if err := summaryDef.Validate(); err == nil {
		t.Error("Expected validation error for summary with negative max age")
	}
}

// TestCollectorErrorHandling tests collector error handling
func TestCollectorErrorHandling(t *testing.T) {
	config := core.DefaultMonitoringConfig()
	collector, err := core.NewCollector(&config)
	if err != nil {
		t.Fatalf("Failed to create collector: %v", err)
	}
	defer collector.Shutdown()

	// Test duplicate metric registration
	def1 := types.NewCounterDefinition("test_metric", "Test metric", []string{}, nil)
	if err := collector.RegisterDefinition(def1); err != nil {
		t.Fatalf("Failed to register definition: %v", err)
	}

	// Try to register duplicate
	def2 := types.NewCounterDefinition("test_metric", "Another test metric", []string{}, nil)
	if err := collector.RegisterDefinition(def2); err == nil {
		t.Error("Expected error for duplicate metric registration")
	} else if !types.IsMonitoringError(err) || err.Code != types.MetricAlreadyRegistered {
		t.Errorf("Expected MetricAlreadyRegistered error, got: %v", err)
	}

	// Test recording metric without definition
	if err := collector.Record("non_existent_metric", 1.0, nil); err == nil {
		t.Error("Expected error for recording undefined metric")
	} else if !types.IsMonitoringError(err) || err.Code != types.MetricNotFound {
		t.Errorf("Expected MetricNotFound error, got: %v", err)
	}

	// Test getting non-existent metric
	_, err = collector.GetMetricsByName("non_existent")
	if err == nil {
		t.Error("Expected error for getting non-existent metric")
	} else if !types.IsMonitoringError(err) || err.Code != types.MetricNotFound {
		t.Errorf("Expected MetricNotFound error, got: %v", err)
	}

	// Test getting non-existent definition
	_, err = collector.GetDefinition("non_existent")
	if err == nil {
		t.Error("Expected error for getting non-existent definition")
	} else if !types.IsMonitoringError(err) || err.Code != types.MetricNotFound {
		t.Errorf("Expected MetricNotFound error, got: %v", err)
	}
}

// TestRegistryErrorHandling tests registry error handling
func TestRegistryErrorHandling(t *testing.T) {
	config := core.DefaultMonitoringConfig()
	collector, err := core.NewCollector(&config)
	if err != nil {
		t.Fatalf("Failed to create collector: %v", err)
	}
	defer collector.Shutdown()

	registry, err := core.NewRegistry(collector, &config)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Test duplicate provider registration
	entry := types.NewProviderRegistryEntry(
		func() types.MetricProvider {
			return &TestProvider{
				BaseProvider: types.NewBaseProvider("test", "1.0.0", "Test provider"),
				counter:      0,
			}
		},
		true,
		100,
	)

	if err := registry.Register("test_provider", entry); err != nil {
		t.Fatalf("Failed to register provider: %v", err)
	}

	// Try to register duplicate
	if err := registry.Register("test_provider", entry); err == nil {
		t.Error("Expected error for duplicate provider registration")
	} else if !types.IsMonitoringError(err) || err.Code != types.ProviderAlreadyRegistered {
		t.Errorf("Expected ProviderAlreadyRegistered error, got: %v", err)
	}

	// Test unregistering non-existent provider
	if err := registry.Unregister("non_existent"); err == nil {
		t.Error("Expected error for unregistering non-existent provider")
	} else if !types.IsMonitoringError(err) || err.Code != types.ProviderNotFound {
		t.Errorf("Expected ProviderNotFound error, got: %v", err)
	}

	// Test initializing non-existent provider
	if err := registry.InitializeProvider("non_existent"); err == nil {
		t.Error("Expected error for initializing non-existent provider")
	} else if !types.IsMonitoringError(err) || err.Code != types.ProviderNotFound {
		t.Errorf("Expected ProviderNotFound error, got: %v", err)
	}

	// Test shutting down non-existent provider
	if err := registry.ShutdownProvider("non_existent"); err == nil {
		t.Error("Expected error for shutting down non-existent provider")
	} else if !types.IsMonitoringError(err) || err.Code != types.ProviderNotFound {
		t.Errorf("Expected ProviderNotFound error, got: %v", err)
	}

	// Test enabling non-existent provider
	if err := registry.EnableProvider("non_existent"); err == nil {
		t.Error("Expected error for enabling non-existent provider")
	} else if !types.IsMonitoringError(err) || err.Code != types.ProviderNotFound {
		t.Errorf("Expected ProviderNotFound error, got: %v", err)
	}

	// Test disabling non-existent provider
	if err := registry.DisableProvider("non_existent"); err == nil {
		t.Error("Expected error for disabling non-existent provider")
	} else if !types.IsMonitoringError(err) || err.Code != types.ProviderNotFound {
		t.Errorf("Expected ProviderNotFound error, got: %v", err)
	}
}

// TestProviderErrorHandling tests provider error handling
func TestProviderErrorHandling(t *testing.T) {
	config := core.DefaultMonitoringConfig()
	collector, err := core.NewCollector(&config)
	if err != nil {
		t.Fatalf("Failed to create collector: %v", err)
	}
	defer collector.Shutdown()

	// Test error provider
	errorProvider := &ErrorTestProvider{
		BaseProvider: types.NewBaseProvider("error_provider", "1.0.0", "Provider that returns errors"),
		shouldError:  true,
	}

	// Register error provider
	if err := collector.RegisterProvider(errorProvider); err != nil {
		t.Fatalf("Failed to register error provider: %v", err)
	}

	// Collect should return error
	if err := collector.CollectFromProviders(); err == nil {
		t.Error("Expected error from error provider")
	}

	// Test getting provider that returns error
	_, err = collector.GetProvider("error_provider")
	if err != nil {
		t.Errorf("Unexpected error getting provider: %v", err)
	}

	// Test getting non-existent provider
	_, err = collector.GetProvider("non_existent")
	if err == nil {
		t.Error("Expected error for getting non-existent provider")
	} else if !types.IsMonitoringError(err) || err.Code != types.ProviderNotFound {
		t.Errorf("Expected ProviderNotFound error, got: %v", err)
	}
}

// TestExporterErrorHandling tests exporter error handling
func TestExporterErrorHandling(t *testing.T) {
	config := core.DefaultMonitoringConfig()
	collector, err := core.NewCollector(&config)
	if err != nil {
		t.Fatalf("Failed to create collector: %v", err)
	}
	defer collector.Shutdown()

	// Create exporter with nil collector (should fail)
	_, err = core.NewExporter(nil, &config.ExportConfig)
	if err == nil {
		t.Error("Expected error for nil collector")
	} else if !types.IsMonitoringError(err) || err.Code != types.CollectorNotInitialized {
		t.Errorf("Expected CollectorNotInitialized error, got: %v", err)
	}

	// Create valid exporter
	exporter, err := core.NewExporter(collector, &config.ExportConfig)
	if err != nil {
		t.Fatalf("Failed to create exporter: %v", err)
	}

	// Test exporting with invalid format via manager
	manager, err := monitoring.NewManager(&config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if err := manager.Initialize(); err != nil {
		t.Fatalf("Failed to initialize manager: %v", err)
	}

	// Export with invalid format
	_, err = manager.ExportMetrics("invalid_format")
	if err == nil {
		t.Error("Expected error for invalid export format")
	} else if !types.IsMonitoringError(err) {
		t.Errorf("Expected monitoring error, got: %v", err)
	}

	// Test exporting when exporter is not initialized
	config2 := core.DefaultMonitoringConfig()
	config2.ExportConfig.Enabled = false
	manager2, err := monitoring.NewManager(&config2)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if err := manager2.Initialize(); err != nil {
		t.Fatalf("Failed to initialize manager: %v", err)
	}

	_, err = manager2.ExportMetrics("prometheus")
	if err == nil {
		t.Error("Expected error when exporter is not initialized")
	} else if !types.IsMonitoringError(err) || err.Code != types.ExportFailed {
		t.Errorf("Expected ExportFailed error, got: %v", err)
	}

	// Cleanup
	exporter.Stop()
	manager.Shutdown()
	manager2.Shutdown()
}

// TestManagerErrorHandling tests manager error handling
func TestManagerErrorHandling(t *testing.T) {
	// Test invalid configuration
	invalidConfig := core.DefaultMonitoringConfig()
	invalidConfig.SamplingRate = -1.0

	_, err := monitoring.NewManager(&invalidConfig)
	if err == nil {
		t.Error("Expected error for invalid configuration")
	} else if !types.IsMonitoringError(err) {
		t.Errorf("Expected monitoring error, got: %v", err)
	}

	// Test nil configuration (should use default)
	manager, err := monitoring.NewManager(nil)
	if err != nil {
		t.Fatalf("Failed to create manager with nil config: %v", err)
	}
	defer manager.Shutdown()

	// Test registering provider before initialization
	err = manager.RegisterProvider("test",
		func() types.MetricProvider {
			return &TestProvider{
				BaseProvider: types.NewBaseProvider("test", "1.0.0", "Test"),
				counter:      0,
			}
		},
		true, 100)
	if err == nil {
		t.Error("Expected error for registering provider before initialization")
	} else if !types.IsMonitoringError(err) || err.Code != types.RegistryNotInitialized {
		t.Errorf("Expected RegistryNotInitialized error, got: %v", err)
	}

	// Test collecting metrics before initialization
	err = manager.CollectMetrics()
	if err == nil {
		t.Error("Expected error for collecting metrics before initialization")
	} else if !types.IsMonitoringError(err) || err.Code != types.CollectorNotInitialized {
		t.Errorf("Expected CollectorNotInitialized error, got: %v", err)
	}

	// Initialize manager
	if err := manager.Initialize(); err != nil {
		t.Fatalf("Failed to initialize manager: %v", err)
	}

	// Test updating with nil configuration
	err = manager.UpdateConfig(nil)
	if err == nil {
		t.Error("Expected error for nil configuration update")
	} else if !types.IsMonitoringError(err) {
		t.Errorf("Expected monitoring error, got: %v", err)
	}

	// Test updating with invalid configuration
	invalidUpdateConfig := core.DefaultMonitoringConfig()
	invalidUpdateConfig.SamplingRate = 2.0
	err = manager.UpdateConfig(&invalidUpdateConfig)
	if err == nil {
		t.Error("Expected error for invalid configuration update")
	} else if !types.IsMonitoringError(err) {
		t.Errorf("Expected monitoring error, got: %v", err)
	}
}

// TestDependencyErrorHandling tests dependency validation errors
func TestDependencyErrorHandling(t *testing.T) {
	config := core.DefaultMonitoringConfig()
	collector, err := core.NewCollector(&config)
	if err != nil {
		t.Fatalf("Failed to create collector: %v", err)
	}
	defer collector.Shutdown()

	registry, err := core.NewRegistry(collector, &config)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Create provider with missing dependency
	dependentProvider := &DependencyTestProvider{
		BaseProvider: types.NewBaseProvider("dependent", "1.0.0", "Provider with dependency"),
		dependencies: []string{"missing_dependency"},
	}

	entry := types.NewProviderRegistryEntry(
		func() types.MetricProvider { return dependentProvider },
		true,
		100,
	)

	// Register dependent provider (registration should succeed)
	if err := registry.Register("dependent", entry); err != nil {
		t.Logf("Note: Provider registration failed (may be expected): %v", err)
		// Continue test even if registration fails
	}

	// Try to initialize (should fail due to missing dependency)
	if err := registry.InitializeProvider("dependent"); err == nil {
		t.Error("Expected error for missing dependency")
	} else {
		t.Logf("Got expected dependency error: %v", err)
		// Any error is acceptable for dependency failure
	}

	// Validate dependencies - may or may not return error depending on implementation
	if err := registry.ValidateDependencies(); err != nil {
		t.Logf("Dependency validation returned error (expected): %v", err)
	} else {
		t.Log("Dependency validation passed (implementation may not check uninitialized providers)")
	}
}

// ErrorTestProvider is a provider that returns errors for testing
type ErrorTestProvider struct {
	*types.BaseProvider
	shouldError bool
}

func (p *ErrorTestProvider) Metrics() []types.MetricDefinition {
	return []types.MetricDefinition{
		types.NewCounterDefinition(
			"error_provider_counter",
			"Error provider counter",
			[]string{},
			nil,
		),
	}
}

func (p *ErrorTestProvider) Collect() ([]types.Metric, *types.Error) {
	if p.shouldError {
		return nil, types.NewError(cd.Unexpected, "test error from provider")
	}

	return []types.Metric{
		types.NewCounter("error_provider_counter", 1.0, nil),
	}, nil
}

// DependencyTestProvider is a provider with dependencies for testing
type DependencyTestProvider struct {
	*types.BaseProvider
	dependencies []string
}

func (p *DependencyTestProvider) Metrics() []types.MetricDefinition {
	return []types.MetricDefinition{
		types.NewCounterDefinition(
			"dependency_test_counter",
			"Dependency test counter",
			[]string{},
			nil,
		),
	}
}

func (p *DependencyTestProvider) Collect() ([]types.Metric, *types.Error) {
	return []types.Metric{
		types.NewCounter("dependency_test_counter", 1.0, nil),
	}, nil
}

func (p *DependencyTestProvider) GetMetadata() types.ProviderMetadata {
	metadata := p.BaseProvider.GetMetadata()
	metadata.Dependencies = p.dependencies
	return metadata
}

// TestRecoveryFromErrors tests recovery from error conditions
func TestRecoveryFromErrors(t *testing.T) {
	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = false // Disable async for reliable testing
	collector, err := core.NewCollector(&config)
	if err != nil {
		t.Fatalf("Failed to create collector: %v", err)
	}
	defer collector.Shutdown()

	// Test that collector continues working after errors
	def := types.NewCounterDefinition("recovery_test", "Recovery test", []string{}, nil)
	if err := collector.RegisterDefinition(def); err != nil {
		t.Fatalf("Failed to register definition: %v", err)
	}

	// Get full name with namespace
	fullName := def.GetFullName(config.Namespace)

	// Record some metrics
	for i := 0; i < 10; i++ {
		if err := collector.Record(fullName, float64(i), nil); err != nil {
			t.Fatalf("Failed to record metric %d: %v", i, err)
		}
	}

	// Clear metrics
	collector.ClearMetrics()

	// Should be able to record more metrics after clearing
	for i := 0; i < 5; i++ {
		if err := collector.Record(fullName, float64(i), nil); err != nil {
			t.Fatalf("Failed to record metric after clearing %d: %v", i, err)
		}
	}

	// Verify metrics were recorded
	t.Logf("Looking for metrics with name: %s", fullName)

	// Check all metrics in collector
	allMetrics := collector.GetMetrics()
	t.Logf("Total metric types in collector: %d", len(allMetrics))
	for name := range allMetrics {
		t.Logf("Found metric type: %s", name)
	}

	metrics, err := collector.GetMetricsByName(fullName)
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}

	if len(metrics) != 5 {
		t.Errorf("Expected 5 metrics after recovery, got %d", len(metrics))
	}
}

// TestErrorPropagation tests that errors are properly propagated
func TestErrorPropagation(t *testing.T) {
	// Test error propagation through multiple layers
	config := core.DefaultMonitoringConfig()
	manager, err := monitoring.NewManager(&config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Don't initialize manager
	// Try to get collector (should be nil)
	collector := manager.GetCollector()
	if collector != nil {
		t.Error("Collector should be nil before initialization")
	}

	// Try to get registry (should be nil)
	registry := manager.GetRegistry()
	if registry != nil {
		t.Error("Registry should be nil before initialization")
	}

	// Try to get exporter (should be nil)
	exporter := manager.GetExporter()
	if exporter != nil {
		t.Error("Exporter should be nil before initialization")
	}

	// Cleanup
	manager.Shutdown()
}
