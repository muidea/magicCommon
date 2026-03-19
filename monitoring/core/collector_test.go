package core

import (
	"testing"
	"time"

	"github.com/muidea/magicCommon/monitoring/types"
)

func TestCollectorRecordKeepsSingleSeriesPerLabelSet(t *testing.T) {
	config := DefaultMonitoringConfig()
	config.AsyncCollection = false

	collector, err := NewCollector(&config)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}

	definition := types.NewGaugeDefinition("test_metric", "test help", []string{"id"}, nil)
	if err := collector.RegisterDefinition(definition); err != nil {
		t.Fatalf("RegisterDefinition() error = %v", err)
	}

	metricName := config.Namespace + "_test_metric"
	labels := map[string]string{"id": "42"}
	if err := collector.Record(metricName, 1, labels); err != nil {
		t.Fatalf("Record() first call error = %v", err)
	}
	if err := collector.Record(metricName, 2, labels); err != nil {
		t.Fatalf("Record() second call error = %v", err)
	}

	metrics, err := collector.GetMetricsByName(metricName)
	if err != nil {
		t.Fatalf("GetMetricsByName() error = %v", err)
	}
	if len(metrics) != 1 {
		t.Fatalf("len(metrics) = %d, want 1", len(metrics))
	}
	if metrics[0].Value != 2 {
		t.Fatalf("metrics[0].Value = %v, want 2", metrics[0].Value)
	}
	if collector.GetMetricCount() != 1 {
		t.Fatalf("GetMetricCount() = %d, want 1", collector.GetMetricCount())
	}
}

func TestCollectorCollectFromProvidersReplacesExistingSeries(t *testing.T) {
	config := DefaultMonitoringConfig()
	config.AsyncCollection = true
	config.BatchSize = 1
	config.BufferSize = 4

	collector, err := NewCollector(&config)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}

	provider := &testMetricProvider{}
	provider.metricsFn = func() []types.Metric {
		providerValue := float64(provider.collectCount)
		provider.collectCount++
		return []types.Metric{
			types.NewGauge(
				"provider_metric",
				providerValue,
				map[string]string{"app": "magicbase"},
			),
		}
	}

	if err := collector.RegisterProvider(provider); err != nil {
		t.Fatalf("RegisterProvider() error = %v", err)
	}

	for idx := 0; idx < 3; idx++ {
		if err := collector.CollectFromProviders(); err != nil {
			t.Fatalf("CollectFromProviders() #%d error = %v", idx+1, err)
		}
	}

	metricName := config.Namespace + "_provider_metric"
	metrics, err := collector.GetMetricsByName(metricName)
	if err != nil {
		t.Fatalf("GetMetricsByName() error = %v", err)
	}
	if len(metrics) != 1 {
		t.Fatalf("len(metrics) = %d, want 1", len(metrics))
	}
	if metrics[0].Value != 2 {
		t.Fatalf("metrics[0].Value = %v, want 2", metrics[0].Value)
	}
	if collector.GetMetricCount() != 1 {
		t.Fatalf("GetMetricCount() = %d, want 1", collector.GetMetricCount())
	}
}

type testMetricProvider struct {
	collectCount int
	metricsFn    func() []types.Metric
}

func (p *testMetricProvider) Name() string {
	return "test-provider"
}

func (p *testMetricProvider) Metrics() []types.MetricDefinition {
	return []types.MetricDefinition{
		types.NewGaugeDefinition("provider_metric", "provider help", []string{"app"}, nil),
	}
}

func (p *testMetricProvider) Init(any) *types.Error {
	return nil
}

func (p *testMetricProvider) Collect() ([]types.Metric, *types.Error) {
	return p.metricsFn(), nil
}

func (p *testMetricProvider) Shutdown() *types.Error {
	return nil
}

func (p *testMetricProvider) GetMetadata() types.ProviderMetadata {
	return types.ProviderMetadata{
		LastCollectionTime: time.Now(),
	}
}
