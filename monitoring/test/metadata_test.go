package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/muidea/magicCommon/monitoring"
	"github.com/muidea/magicCommon/monitoring/core"
	"github.com/muidea/magicCommon/monitoring/types"
	"github.com/stretchr/testify/assert"
)

// MetadataTestProvider is a test provider with metadata
type MetadataTestProvider struct {
	*types.BaseProvider
	counter int
}

func NewMetadataTestProvider() *MetadataTestProvider {
	return &MetadataTestProvider{
		BaseProvider: types.NewBaseProvider("metadata_test", "1.0.0", "Test metadata provider"),
		counter:      0,
	}
}

func (p *MetadataTestProvider) Metrics() []types.MetricDefinition {
	return []types.MetricDefinition{
		types.NewCounterDefinition(
			"metadata_test_requests_total",
			"Total number of metadata test requests",
			[]string{"method", "endpoint"},
			map[string]string{"environment": "test"},
		),
		types.NewGaugeDefinition(
			"metadata_test_response_time_seconds",
			"Response time in seconds",
			[]string{"endpoint"},
			nil,
		),
		types.NewHistogramDefinition(
			"metadata_test_request_size_bytes",
			"Request size in bytes",
			[]string{"method"},
			[]float64{100, 500, 1000, 5000, 10000},
			nil,
		),
	}
}

func (p *MetadataTestProvider) Collect() ([]types.Metric, *types.Error) {
	p.counter++

	return []types.Metric{
		types.NewCounter(
			"metadata_test_requests_total",
			float64(p.counter),
			map[string]string{
				"method":   "GET",
				"endpoint": "/api/test",
			},
		),
		types.NewGauge(
			"metadata_test_response_time_seconds",
			0.125,
			map[string]string{
				"endpoint": "/api/test",
			},
		),
	}, nil
}

func ensureMonitoringServerReachable(t *testing.T, port int) {
	t.Helper()

	client := &http.Client{Timeout: 200 * time.Millisecond}
	baseURL := fmt.Sprintf("http://localhost:%d/", port)
	deadline := time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {
		resp, err := client.Get(baseURL)
		if err == nil {
			_ = resp.Body.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Skipf("monitoring exporter is not reachable on port %d in current environment", port)
}

func TestMetadataEndpoints(t *testing.T) {
	// Create a custom configuration with metadata endpoints enabled
	config := core.DefaultMonitoringConfig()
	config.ExportConfig.Enabled = true
	config.ExportConfig.Port = 9091 // Use different port to avoid conflicts
	config.ExportConfig.EnablePrometheus = true
	config.ExportConfig.EnableJSON = true
	config.ExportConfig.MetadataPath = "/api/metadata"

	// Create manager
	manager, err := monitoring.NewManager(&config)
	assert.Nil(t, err, "Failed to create manager")

	// Initialize manager
	err = manager.Initialize()
	assert.Nil(t, err, "Failed to initialize manager")

	// Start manager
	err = manager.Start()
	assert.Nil(t, err, "Failed to start manager")
	defer func() { _ = manager.Shutdown() }()

	// Register test provider
	err = manager.RegisterProvider(
		"metadata_test",
		func() types.MetricProvider { return NewMetadataTestProvider() },
		true, // autoInitialize
		100,  // priority
	)
	assert.Nil(t, err, "Failed to register provider")

	// Give some time for server to start
	time.Sleep(100 * time.Millisecond)
	ensureMonitoringServerReachable(t, config.ExportConfig.Port)

	// Test 1: Get all metadata
	t.Run("GetAllMetadata", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d%s", config.ExportConfig.Port, config.ExportConfig.MetadataPath))
		if !assert.Nil(t, err, "Failed to get metadata") {
			return
		}
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200")
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Expected JSON content type")

		var result struct {
			Timestamp time.Time                       `json:"timestamp"`
			Metrics   map[string]types.MetricMetadata `json:"metrics"`
			Count     int                             `json:"count"`
		}

		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.Nil(t, err, "Failed to decode response")
		assert.True(t, result.Count >= 3, "Expected at least 3 metrics")

		// Check for metrics with namespace prefix
		foundRequests := false
		foundResponseTime := false
		foundRequestSize := false
		var requestsMetric types.MetricMetadata

		for name, metric := range result.Metrics {
			if strings.HasSuffix(name, "metadata_test_requests_total") {
				foundRequests = true
				requestsMetric = metric
			}
			if strings.HasSuffix(name, "metadata_test_response_time_seconds") {
				foundResponseTime = true
			}
			if strings.HasSuffix(name, "metadata_test_request_size_bytes") {
				foundRequestSize = true
			}
		}

		assert.True(t, foundRequests, "Should find metadata_test_requests_total")
		assert.True(t, foundResponseTime, "Should find metadata_test_response_time_seconds")
		assert.True(t, foundRequestSize, "Should find metadata_test_request_size_bytes")

		// Verify metadata structure
		assert.Equal(t, types.CounterMetric, requestsMetric.Type)
		assert.Equal(t, "Total number of metadata test requests", requestsMetric.Help)
		assert.Equal(t, []string{"method", "endpoint"}, requestsMetric.LabelNames)
		assert.Equal(t, types.QualityMedium, requestsMetric.Quality) // Default quality
	})

	// Test 2: Get single metric metadata
	t.Run("GetSingleMetricMetadata", func(t *testing.T) {
		// First get all metadata to find the full metric name
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d%s", config.ExportConfig.Port, config.ExportConfig.MetadataPath))
		if !assert.Nil(t, err, "Failed to get all metadata") {
			return
		}
		defer func() { _ = resp.Body.Close() }()

		var allMetadata struct {
			Metrics map[string]types.MetricMetadata `json:"metrics"`
		}
		err = json.NewDecoder(resp.Body).Decode(&allMetadata)
		assert.Nil(t, err, "Failed to decode all metadata response")

		// Find the metric with suffix "metadata_test_requests_total"
		var fullMetricName string
		for name := range allMetadata.Metrics {
			if strings.HasSuffix(name, "metadata_test_requests_total") {
				fullMetricName = name
				break
			}
		}
		assert.NotEmpty(t, fullMetricName, "Should find metric with suffix metadata_test_requests_total")

		// Now get single metric metadata
		resp, err = http.Get(fmt.Sprintf("http://localhost:%d%s/%s", config.ExportConfig.Port, config.ExportConfig.MetadataPath, fullMetricName))
		if !assert.Nil(t, err, "Failed to get single metric metadata") {
			return
		}
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200")
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Expected JSON content type")

		var metadata types.MetricMetadata
		err = json.NewDecoder(resp.Body).Decode(&metadata)
		assert.Nil(t, err, "Failed to decode response")

		assert.Equal(t, fullMetricName, metadata.Name)
		assert.Equal(t, types.CounterMetric, metadata.Type)
		assert.Equal(t, "Total number of metadata test requests", metadata.Help)
		assert.Equal(t, []string{"method", "endpoint"}, metadata.LabelNames)
	})

	// Test 3: Get non-existent metric metadata
	t.Run("GetNonExistentMetricMetadata", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d%s/nonexistent_metric", config.ExportConfig.Port, config.ExportConfig.MetadataPath))
		if !assert.Nil(t, err, "Failed to get non-existent metric metadata") {
			return
		}
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Expected status 404 for non-existent metric")
	})

	// Test 4: Verify metadata endpoint is listed in info
	t.Run("MetadataEndpointInInfo", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/", config.ExportConfig.Port))
		if !assert.Nil(t, err, "Failed to get info") {
			return
		}
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200")

		var info struct {
			Endpoints []string `json:"endpoints"`
		}
		err = json.NewDecoder(resp.Body).Decode(&info)
		assert.Nil(t, err, "Failed to decode info response")

		assert.Contains(t, info.Endpoints, config.ExportConfig.MetadataPath, "Metadata endpoint should be listed in info")
	})
}

func TestMetadataValidation(t *testing.T) {
	// Test MetricMetadata validation
	t.Run("MetricMetadataValidation", func(t *testing.T) {
		// Valid metadata
		validMetadata := types.MetricMetadata{
			MetricDefinition: types.NewCounterDefinition(
				"valid_metric",
				"A valid metric",
				[]string{"label1"},
				nil,
			),
			Unit:        "requests",
			Aggregation: "sum",
			Quality:     types.QualityHigh,
		}
		err := validMetadata.Validate()
		assert.Nil(t, err, "Valid metadata should pass validation")

		// Invalid aggregation
		invalidAggMetadata := validMetadata
		invalidAggMetadata.Aggregation = "invalid_agg"
		err = invalidAggMetadata.Validate()
		assert.NotNil(t, err, "Invalid aggregation should fail validation")
		assert.Contains(t, err.Error(), "invalid aggregation")

		// Invalid alert thresholds (warning >= critical)
		invalidThresholdMetadata := validMetadata
		invalidThresholdMetadata.Aggregation = "" // Clear invalid aggregation
		invalidThresholdMetadata.AlertThreshold = &types.AlertThreshold{
			Warning:  100,
			Critical: 50, // Warning > Critical
		}
		err = invalidThresholdMetadata.Validate()
		assert.NotNil(t, err, "Invalid alert thresholds should fail validation")
		assert.Contains(t, err.Error(), "warning threshold must be less than critical")
	})

	t.Run("NewMetricMetadataFunction", func(t *testing.T) {
		def := types.NewCounterDefinition(
			"test_metric",
			"Test metric",
			[]string{"label"},
			nil,
		)

		metadata := types.NewMetricMetadata(def)
		assert.Equal(t, def.Name, metadata.Name)
		assert.Equal(t, def.Type, metadata.Type)
		assert.Equal(t, def.Help, metadata.Help)
		assert.Equal(t, types.QualityMedium, metadata.Quality) // Default quality
	})
}
