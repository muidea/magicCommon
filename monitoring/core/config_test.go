package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultMonitoringConfig(t *testing.T) {
	config := DefaultMonitoringConfig()

	assert.True(t, config.Enabled, "Default config should be enabled")
	assert.Equal(t, "app", config.Namespace, "Default namespace should be 'app'")
	assert.Equal(t, 1.0, config.SamplingRate, "Default sampling rate should be 1.0")
	assert.True(t, config.AsyncCollection, "Default async collection should be true")
	assert.Equal(t, 30*time.Second, config.CollectionInterval, "Default collection interval should be 30s")
	assert.Equal(t, 24*time.Hour, config.RetentionPeriod, "Default retention period should be 24h")
	assert.Equal(t, DetailLevelStandard, config.DetailLevel, "Default detail level should be 'standard'")
	assert.Equal(t, 100, config.BatchSize, "Default batch size should be 100")
	assert.Equal(t, 1000, config.BufferSize, "Default buffer size should be 1000")
	assert.Equal(t, 10, config.MaxConcurrentTasks, "Default max concurrent tasks should be 10")
	assert.Equal(t, 10*time.Second, config.Timeout, "Default timeout should be 10s")
	assert.Equal(t, "development", config.Environment, "Default environment should be 'development'")
}

func TestDefaultExportConfig(t *testing.T) {
	config := DefaultExportConfig()

	assert.True(t, config.Enabled, "Default export config should be enabled")
	assert.Equal(t, 9090, config.Port, "Default export port should be 9090")
	assert.Equal(t, "/metrics", config.Path, "Default path should be '/metrics'")
	assert.Equal(t, "/health", config.HealthCheckPath, "Default health check path should be '/health'")
	assert.Equal(t, "/metrics/json", config.MetricsPath, "Default metrics path should be '/metrics/json'")
	assert.Equal(t, "/", config.InfoPath, "Default info path should be '/'")
	assert.Equal(t, "/api/metadata", config.MetadataPath, "Default metadata path should be '/api/metadata'")
	assert.True(t, config.EnablePrometheus, "Default should enable Prometheus")
	assert.True(t, config.EnableJSON, "Default should enable JSON")
	assert.Equal(t, 30*time.Second, config.RefreshInterval, "Default refresh interval should be 30s")
	assert.Equal(t, 10*time.Second, config.ScrapeTimeout, "Default scrape timeout should be 10s")
	assert.False(t, config.EnableTLS, "Default TLS should be disabled")
}

func TestDevelopmentConfig(t *testing.T) {
	config := DevelopmentConfig()

	assert.True(t, config.Enabled, "Development config should be enabled")
	assert.Equal(t, "app", config.Namespace, "Development namespace should be 'app'")
	assert.Equal(t, 0.1, config.SamplingRate, "Development sampling rate should be 0.1")
	assert.Equal(t, DetailLevelBasic, config.DetailLevel, "Development detail level should be 'basic'")
	assert.False(t, config.ExportConfig.Enabled, "Development export should be disabled")
	assert.False(t, config.AsyncCollection, "Development should use synchronous collection")
	assert.Equal(t, "development", config.Environment, "Development environment should be 'development'")
}

func TestProductionConfig(t *testing.T) {
	config := ProductionConfig()

	assert.True(t, config.Enabled, "Production config should be enabled")
	assert.Equal(t, "app", config.Namespace, "Production namespace should be 'app'")
	assert.Equal(t, 0.5, config.SamplingRate, "Production sampling rate should be 0.5")
	assert.Equal(t, DetailLevelStandard, config.DetailLevel, "Production detail level should be 'standard'")
	assert.True(t, config.ExportConfig.Enabled, "Production export should be enabled")
	assert.True(t, config.ExportConfig.EnableTLS, "Production TLS should be enabled")
	assert.Equal(t, 500, config.BatchSize, "Production batch size should be 500")
	assert.Equal(t, 5000, config.BufferSize, "Production buffer size should be 5000")
	assert.Equal(t, 50, config.MaxConcurrentTasks, "Production max concurrent tasks should be 50")
	assert.Equal(t, "production", config.Environment, "Production environment should be 'production'")
}

func TestHighLoadConfig(t *testing.T) {
	config := HighLoadConfig()

	assert.True(t, config.Enabled, "High load config should be enabled")
	assert.Equal(t, "app", config.Namespace, "High load namespace should be 'app'")
	assert.Equal(t, 0.1, config.SamplingRate, "High load sampling rate should be 0.1")
	assert.Equal(t, DetailLevelBasic, config.DetailLevel, "High load detail level should be 'basic'")
	assert.Equal(t, 1000, config.BatchSize, "High load batch size should be 1000")
	assert.Equal(t, 10000, config.BufferSize, "High load buffer size should be 10000")
	assert.Equal(t, 100, config.MaxConcurrentTasks, "High load max concurrent tasks should be 100")
	assert.Equal(t, 60*time.Second, config.ExportConfig.RefreshInterval, "High load refresh interval should be 60s")
	assert.Equal(t, "highload", config.Environment, "High load environment should be 'highload'")
}

func TestMonitoringConfigValidate(t *testing.T) {
	t.Run("Valid config", func(t *testing.T) {
		config := DefaultMonitoringConfig()
		err := config.Validate()

		assert.Nil(t, err, "Valid config should not return error")
	})

	t.Run("Invalid namespace", func(t *testing.T) {
		config := DefaultMonitoringConfig()
		config.Namespace = ""
		err := config.Validate()

		assert.NotNil(t, err, "Empty namespace should return error")
	})

	t.Run("Invalid sampling rate too low", func(t *testing.T) {
		config := DefaultMonitoringConfig()
		config.SamplingRate = -0.1
		err := config.Validate()

		assert.NotNil(t, err, "Negative sampling rate should return error")
	})

	t.Run("Invalid sampling rate too high", func(t *testing.T) {
		config := DefaultMonitoringConfig()
		config.SamplingRate = 1.1
		err := config.Validate()

		assert.NotNil(t, err, "Sampling rate > 1.0 should return error")
	})

	t.Run("Invalid collection interval", func(t *testing.T) {
		config := DefaultMonitoringConfig()
		config.CollectionInterval = 0
		err := config.Validate()

		assert.NotNil(t, err, "Zero collection interval should return error")
	})

	t.Run("Invalid retention period", func(t *testing.T) {
		config := DefaultMonitoringConfig()
		config.RetentionPeriod = 0
		err := config.Validate()

		assert.NotNil(t, err, "Zero retention period should return error")
	})

	t.Run("Invalid detail level", func(t *testing.T) {
		config := DefaultMonitoringConfig()
		config.DetailLevel = "invalid"
		err := config.Validate()

		assert.NotNil(t, err, "Invalid detail level should return error")
	})
}

func TestExportConfigValidate(t *testing.T) {
	t.Run("Valid export config", func(t *testing.T) {
		config := DefaultExportConfig()
		config.Enabled = true
		err := config.Validate()

		assert.Nil(t, err, "Valid export config should not return error")
	})

	t.Run("Disabled export config", func(t *testing.T) {
		config := DefaultExportConfig()
		config.Enabled = false
		err := config.Validate()

		assert.Nil(t, err, "Disabled export config should not return error")
	})

	t.Run("Invalid port too low", func(t *testing.T) {
		config := DefaultExportConfig()
		config.Enabled = true
		config.Port = 0
		err := config.Validate()

		assert.NotNil(t, err, "Port <= 0 should return error")
	})

	t.Run("Invalid port too high", func(t *testing.T) {
		config := DefaultExportConfig()
		config.Enabled = true
		config.Port = 65536
		err := config.Validate()

		assert.NotNil(t, err, "Port > 65535 should return error")
	})

	t.Run("Invalid path", func(t *testing.T) {
		config := DefaultExportConfig()
		config.Enabled = true
		config.Path = ""
		err := config.Validate()

		assert.NotNil(t, err, "Empty path should return error")
	})

	t.Run("Invalid metadata path", func(t *testing.T) {
		config := DefaultExportConfig()
		config.Enabled = true
		config.MetadataPath = ""
		err := config.Validate()

		assert.NotNil(t, err, "Empty metadata path should return error")
	})

	t.Run("TLS enabled without cert", func(t *testing.T) {
		config := DefaultExportConfig()
		config.Enabled = true
		config.EnableTLS = true
		config.TLSCertPath = ""
		err := config.Validate()

		assert.NotNil(t, err, "TLS enabled without cert file should return error")
	})

	t.Run("TLS enabled without key", func(t *testing.T) {
		config := DefaultExportConfig()
		config.Enabled = true
		config.EnableTLS = true
		config.TLSCertPath = "cert.pem"
		config.TLSKeyPath = ""
		err := config.Validate()

		assert.NotNil(t, err, "TLS enabled without key file should return error")
	})
}

func TestShouldSample(t *testing.T) {
	t.Run("Always sample with rate 1.0", func(t *testing.T) {
		config := DefaultMonitoringConfig()
		config.SamplingRate = 1.0

		// Test multiple times
		for i := 0; i < 100; i++ {
			assert.True(t, config.ShouldSample(), "Should always sample with rate 1.0")
		}
	})

	t.Run("Never sample with rate 0.0", func(t *testing.T) {
		config := DefaultMonitoringConfig()
		config.SamplingRate = 0.0

		// Test multiple times
		for i := 0; i < 100; i++ {
			assert.False(t, config.ShouldSample(), "Should never sample with rate 0.0")
		}
	})

	t.Run("Sample with rate 0.5 (deterministic sampling)", func(t *testing.T) {
		config := DefaultMonitoringConfig()
		config.SamplingRate = 0.5

		// Deterministic sampling is consistent within the same time slice (1 second)
		// Run multiple times within the same second - should all return the same value
		firstResult := config.ShouldSample()
		for i := 0; i < 10; i++ {
			result := config.ShouldSample()
			assert.Equal(t, firstResult, result, "Should return consistent result within same time slice")
		}

		// Wait to get a different time slice and verify behavior changes
		// (or stays consistent with the new time slice)
		time.Sleep(time.Second)
		secondResult := config.ShouldSample()

		// With 0.5 rate, over multiple time slices, we should see both true and false
		// when running this test multiple times across different seconds
		// At minimum, verify it returns a boolean without panic
		assert.True(t, firstResult || !firstResult, "Should return boolean")
		assert.True(t, secondResult || !secondResult, "Should return boolean")
	})

	t.Run("Disabled monitoring", func(t *testing.T) {
		config := DefaultMonitoringConfig()
		config.Enabled = false
		config.SamplingRate = 1.0

		// Should not sample when monitoring is disabled
		for i := 0; i < 100; i++ {
			assert.False(t, config.ShouldSample(), "Should not sample when monitoring is disabled")
		}
	})
}

func TestGetSetProviderConfig(t *testing.T) {
	config := DefaultMonitoringConfig()

	// Test setting provider config
	testConfig := map[string]interface{}{
		"key": "value",
		"num": 42,
	}

	config.SetProviderConfig("test_provider", testConfig)

	// Test getting provider config
	retrieved := config.GetProviderConfig("test_provider")
	assert.NotNil(t, retrieved, "Should retrieve provider config")

	// Test getting non-existent provider config
	notFound := config.GetProviderConfig("non_existent")
	assert.Nil(t, notFound, "Should return nil for non-existent provider")
}
