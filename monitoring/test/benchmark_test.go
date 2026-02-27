package test

import (
	"testing"
	"time"

	"github.com/muidea/magicCommon/monitoring"
	"github.com/muidea/magicCommon/monitoring/core"
	"github.com/muidea/magicCommon/monitoring/types"
)

// BenchmarkCollectorRecord benchmarks metric recording performance
func BenchmarkCollectorRecord(b *testing.B) {
	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = false // Disable async for accurate benchmarking
	collector, err := core.NewCollector(&config)
	if err != nil {
		b.Fatalf("Failed to create collector: %v", err)
	}
	defer func() { _ = collector.Shutdown() }()

	// Register a test metric
	def := types.NewCounterDefinition(
		"benchmark_counter",
		"Benchmark counter",
		[]string{"label1", "label2"},
		nil,
	)
	if err := collector.RegisterDefinition(def); err != nil {
		b.Fatalf("Failed to register definition: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		labels := map[string]string{
			"label1": "value1",
			"label2": "value2",
		}
		if err := collector.Record("benchmark_counter", float64(i), labels); err != nil {
			b.Fatalf("Failed to record metric: %v", err)
		}
	}
}

// BenchmarkCollectorRecordAsync benchmarks async metric recording
func BenchmarkCollectorRecordAsync(b *testing.B) {
	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = true
	config.BatchSize = 1000
	collector, err := core.NewCollector(&config)
	if err != nil {
		b.Fatalf("Failed to create collector: %v", err)
	}
	defer func() { _ = collector.Shutdown() }()

	// Register a test metric
	def := types.NewCounterDefinition(
		"benchmark_counter_async",
		"Benchmark async counter",
		[]string{"label"},
		nil,
	)
	if err := collector.RegisterDefinition(def); err != nil {
		b.Fatalf("Failed to register definition: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		labels := map[string]string{
			"label": "value",
		}
		if err := collector.Record("benchmark_counter_async", float64(i), labels); err != nil {
			b.Fatalf("Failed to record metric: %v", err)
		}
	}

	// Force flush to ensure all metrics are processed
	_ = collector.ForceFlush()
}

// BenchmarkExporterPrometheus benchmarks Prometheus format export
func BenchmarkExporterPrometheus(b *testing.B) {
	config := core.DefaultMonitoringConfig()
	collector, err := core.NewCollector(&config)
	if err != nil {
		b.Fatalf("Failed to create collector: %v", err)
	}
	defer func() { _ = collector.Shutdown() }()

	// Register and record some metrics
	for i := 0; i < 1000; i++ {
		def := types.NewCounterDefinition(
			"benchmark_metric_"+string(rune('a'+i%26)),
			"Benchmark metric",
			[]string{"index"},
			nil,
		)
		if err := collector.RegisterDefinition(def); err != nil {
			b.Fatalf("Failed to register definition: %v", err)
		}

		labels := map[string]string{"index": string(rune('a' + i%26))}
		if err := collector.Record(def.Name, float64(i), labels); err != nil {
			b.Fatalf("Failed to record metric: %v", err)
		}
	}

	exporter, err := core.NewExporter(collector, &config.ExportConfig)
	if err != nil {
		b.Fatalf("Failed to create exporter: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := exporter.ExportPrometheus()
		if err != nil {
			b.Fatalf("Failed to export Prometheus metrics: %v", err)
		}
	}
}

// BenchmarkExporterJSON benchmarks JSON format export
func BenchmarkExporterJSON(b *testing.B) {
	config := core.DefaultMonitoringConfig()
	collector, err := core.NewCollector(&config)
	if err != nil {
		b.Fatalf("Failed to create collector: %v", err)
	}
	defer func() { _ = collector.Shutdown() }()

	// Register and record some metrics
	for i := 0; i < 1000; i++ {
		def := types.NewCounterDefinition(
			"benchmark_metric_"+string(rune('a'+i%26)),
			"Benchmark metric",
			[]string{"index"},
			nil,
		)
		if err := collector.RegisterDefinition(def); err != nil {
			b.Fatalf("Failed to register definition: %v", err)
		}

		labels := map[string]string{"index": string(rune('a' + i%26))}
		if err := collector.Record(def.Name, float64(i), labels); err != nil {
			b.Fatalf("Failed to record metric: %v", err)
		}
	}

	exporter, err := core.NewExporter(collector, &config.ExportConfig)
	if err != nil {
		b.Fatalf("Failed to create exporter: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := exporter.ExportJSON()
		if err != nil {
			b.Fatalf("Failed to export JSON metrics: %v", err)
		}
	}
}

// BenchmarkManagerFullCycle benchmarks complete manager lifecycle
func BenchmarkManagerFullCycle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		manager, err := monitoring.NewDevelopmentManager()
		if err != nil {
			b.Fatalf("Failed to create manager: %v", err)
		}

		if err := manager.Initialize(); err != nil {
			b.Fatalf("Failed to initialize manager: %v", err)
		}

		if err := manager.Start(); err != nil {
			b.Fatalf("Failed to start manager: %v", err)
		}

		// Register a test provider
		if err := manager.RegisterProvider(
			"benchmark_provider",
			func() types.MetricProvider { return NewTestProvider() },
			true,
			100,
		); err != nil {
			b.Fatalf("Failed to register provider: %v", err)
		}

		// Collect metrics
		if err := manager.CollectMetrics(); err != nil {
			b.Fatalf("Failed to collect metrics: %v", err)
		}

		// Export metrics
		_, err = manager.ExportMetrics("prometheus")
		if err != nil {
			// This might fail if export is disabled in development config
			// Just log and continue
			b.Logf("Export failed (expected in development): %v", err)
		}

		if err := manager.Shutdown(); err != nil {
			b.Fatalf("Failed to shutdown manager: %v", err)
		}
	}
}

// BenchmarkConcurrentRecording benchmarks concurrent metric recording
func BenchmarkConcurrentRecording(b *testing.B) {
	config := core.DefaultMonitoringConfig()
	config.AsyncCollection = true
	config.BatchSize = 10000
	config.BufferSize = 100000
	collector, err := core.NewCollector(&config)
	if err != nil {
		b.Fatalf("Failed to create collector: %v", err)
	}
	defer func() { _ = collector.Shutdown() }()

	// Register test metrics
	for i := 0; i < 10; i++ {
		def := types.NewCounterDefinition(
			"concurrent_metric_"+string(rune('a'+i)),
			"Concurrent metric",
			[]string{"worker", "iteration"},
			nil,
		)
		if err := collector.RegisterDefinition(def); err != nil {
			b.Fatalf("Failed to register definition: %v", err)
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		workerID := time.Now().UnixNano()
		iteration := 0
		for pb.Next() {
			metricName := "concurrent_metric_" + string(rune('a'+(iteration%10)))
			labels := map[string]string{
				"worker":    string(rune('0' + workerID%10)),
				"iteration": string(rune('0' + iteration%10)),
			}
			if err := collector.Record(metricName, float64(iteration), labels); err != nil {
				b.Fatalf("Failed to record metric: %v", err)
			}
			iteration++
		}
	})

	// Force flush to ensure all metrics are processed
	_ = collector.ForceFlush()
}

// BenchmarkProviderCollection benchmarks provider metric collection
func BenchmarkProviderCollection(b *testing.B) {
	config := core.DefaultMonitoringConfig()
	collector, err := core.NewCollector(&config)
	if err != nil {
		b.Fatalf("Failed to create collector: %v", err)
	}
	defer func() { _ = collector.Shutdown() }()

	// Create a benchmark provider
	provider := &BenchmarkProvider{
		BaseProvider: types.NewBaseProvider("benchmark", "1.0.0", "Benchmark provider"),
		metricCount:  100,
	}

	// Register provider
	if err := collector.RegisterProvider(provider); err != nil {
		b.Fatalf("Failed to register provider: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := collector.CollectFromProviders(); err != nil {
			b.Fatalf("Failed to collect from providers: %v", err)
		}
	}
}

// BenchmarkProvider is a provider for benchmarking
type BenchmarkProvider struct {
	*types.BaseProvider
	metricCount int
	counter     int64
}

func (p *BenchmarkProvider) Metrics() []types.MetricDefinition {
	defs := make([]types.MetricDefinition, p.metricCount)
	for i := 0; i < p.metricCount; i++ {
		defs[i] = types.NewCounterDefinition(
			"benchmark_provider_metric_"+string(rune('a'+i%26)),
			"Benchmark provider metric",
			[]string{"index", "type"},
			nil,
		)
	}
	return defs
}

func (p *BenchmarkProvider) Collect() ([]types.Metric, *types.Error) {
	p.counter++
	metrics := make([]types.Metric, p.metricCount)
	for i := 0; i < p.metricCount; i++ {
		metrics[i] = types.NewCounter(
			"benchmark_provider_metric_"+string(rune('a'+i%26)),
			float64(p.counter),
			map[string]string{
				"index": string(rune('a' + i%26)),
				"type":  "benchmark",
			},
		)
	}
	return metrics, nil
}
