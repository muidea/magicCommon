package test

import (
	"testing"

	"github.com/muidea/magicCommon/monitoring/core"
	"github.com/muidea/magicCommon/monitoring/types"
)

// TestQuickCollector tests basic collector functionality
func TestQuickCollector(t *testing.T) {
	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = false // Use sync mode for testing

	collector, err := core.NewCollector(&config)
	if err != nil {
		t.Fatalf("Failed to create collector: %v", err)
	}
	defer func() { _ = collector.Shutdown() }()

	// Test metric definition
	def := types.NewCounterDefinition(
		"test_counter",
		"Test counter metric",
		[]string{"label1"},
		nil,
	)

	if err := collector.RegisterDefinition(def); err != nil {
		t.Fatalf("Failed to register definition: %v", err)
	}

	// Test recording metrics - use full name with namespace
	fullName := def.GetFullName(config.Namespace)
	for i := 0; i < 5; i++ {
		labels := map[string]string{"label1": "value1"}
		if err := collector.Record(fullName, float64(i), labels); err != nil {
			t.Fatalf("Failed to record metric %d: %v", i, err)
		}
	}

	// Test getting metrics
	metrics, err := collector.GetMetricsByName(fullName)
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}

	if len(metrics) != 1 {
		t.Errorf("Expected 1 stored series, got %d", len(metrics))
	}

	if len(metrics) == 1 && metrics[0].Value != 4 {
		t.Errorf("Expected latest metric value 4, got %v", metrics[0].Value)
	}

	// Test stats
	stats := collector.GetStats()
	if stats.MetricsCollected != 5 {
		t.Errorf("Expected 5 metrics collected, got %d", stats.MetricsCollected)
	}

	t.Logf("Test passed: recorded %d samples and retained %d series", stats.MetricsCollected, len(metrics))
}

// TestQuickRegistry tests basic registry functionality
func TestQuickRegistry(t *testing.T) {
	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = false // Use sync mode for testing

	collector, err := core.NewCollector(&config)
	if err != nil {
		t.Fatalf("Failed to create collector: %v", err)
	}
	defer func() { _ = collector.Shutdown() }()

	registry, err := core.NewRegistry(collector, &config)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Create a simple provider
	provider := &QuickTestProvider{
		BaseProvider: types.NewBaseProvider("quick", "1.0.0", "Quick test provider"),
	}

	// Register provider factory
	entry := types.NewProviderRegistryEntry(
		func() types.MetricProvider { return provider },
		true, // autoInitialize - now should work with deadlock fix
		100,  // priority
	)

	if err := registry.Register("quick_provider", entry); err != nil {
		t.Fatalf("Failed to register provider: %v", err)
	}

	// Check provider was registered
	stats := registry.GetStats()
	if stats.TotalProvidersRegistered != 1 {
		t.Errorf("Expected 1 provider registered, got %d", stats.TotalProvidersRegistered)
	}

	// Skip initialization for now - just test registration
	t.Logf("Test passed: provider registered successfully")
}

// TestQuickConfigValidation tests configuration validation
func TestQuickConfigValidation(t *testing.T) {
	// Test valid config
	validConfig := core.DefaultMonitoringConfig()
	if err := validConfig.Validate(); err != nil {
		t.Errorf("Valid config should pass validation: %v", err)
	}

	// Test invalid sampling rate
	invalidConfig := core.DefaultMonitoringConfig()
	invalidConfig.SamplingRate = 1.5
	if err := invalidConfig.Validate(); err == nil {
		t.Error("Invalid sampling rate should fail validation")
	}

	// Test valid environment configs
	devConfig := core.DevelopmentConfig()
	if err := devConfig.Validate(); err != nil {
		t.Errorf("Development config should be valid: %v", err)
	}

	prodConfig := core.ProductionConfig()
	prodConfig.ExportConfig.EnableTLS = false // Disable TLS for testing
	if err := prodConfig.Validate(); err != nil {
		t.Errorf("Production config should be valid: %v", err)
	}

	highLoadConfig := core.HighLoadConfig()
	highLoadConfig.ExportConfig.EnableTLS = false // Disable TLS for testing
	if err := highLoadConfig.Validate(); err != nil {
		t.Errorf("High load config should be valid: %v", err)
	}

	t.Log("Test passed: configuration validation works")
}

// QuickTestProvider is a simple test provider
type QuickTestProvider struct {
	*types.BaseProvider
	counter int
}

func (p *QuickTestProvider) Metrics() []types.MetricDefinition {
	return []types.MetricDefinition{
		types.NewCounterDefinition(
			"quick_counter",
			"Quick test counter",
			[]string{"test"},
			nil,
		),
	}
}

func (p *QuickTestProvider) Collect() ([]types.Metric, *types.Error) {
	p.counter++
	return []types.Metric{
		types.NewCounter(
			"quick_counter",
			float64(p.counter),
			map[string]string{"test": "quick"},
		),
	}, nil
}

// TestQuickErrorHandling tests basic error handling
func TestQuickErrorHandling(t *testing.T) {
	config := core.DefaultMonitoringConfig()
	collector, err := core.NewCollector(&config)
	if err != nil {
		t.Fatalf("Failed to create collector: %v", err)
	}
	defer func() { _ = collector.Shutdown() }()

	// Test duplicate metric registration
	def1 := types.NewCounterDefinition("duplicate", "Test", []string{}, nil)
	if err := collector.RegisterDefinition(def1); err != nil {
		t.Fatalf("Failed to register first definition: %v", err)
	}

	def2 := types.NewCounterDefinition("duplicate", "Test again", []string{}, nil)
	if err := collector.RegisterDefinition(def2); err == nil {
		t.Error("Duplicate registration should fail")
	} else if !types.IsMonitoringError(err) {
		t.Errorf("Expected monitoring error, got: %v", err)
	}

	// Test recording undefined metric
	if err := collector.Record("undefined_metric", 1.0, nil); err == nil {
		t.Error("Recording undefined metric should fail")
	}

	t.Log("Test passed: error handling works")
}
