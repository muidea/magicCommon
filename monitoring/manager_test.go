package monitoring

import (
	"testing"

	"github.com/muidea/magicCommon/monitoring/core"
	"github.com/muidea/magicCommon/monitoring/types"
)

type testMetricProvider struct {
	*types.BaseProvider
}

func newTestMetricProvider(name string) *testMetricProvider {
	return &testMetricProvider{
		BaseProvider: types.NewBaseProvider(name, "1.0.0", "test provider"),
	}
}

func (p *testMetricProvider) Metrics() []types.MetricDefinition {
	return nil
}

func TestManagerRegisterProviderUsesInstanceRegistry(t *testing.T) {
	config := core.DefaultMonitoringConfig()
	config.ExportConfig.Enabled = false

	manager, err := NewManager(&config)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	if err = manager.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if err = manager.RegisterProvider("local-test", func() types.MetricProvider {
		return newTestMetricProvider("local-test")
	}, true, 10); err != nil {
		t.Fatalf("RegisterProvider failed: %v", err)
	}

	if _, ok := manager.GetRegistry().GetRegistryEntries()["local-test"]; !ok {
		t.Fatal("provider was not registered in manager registry")
	}
}

func TestManagerUpdateConfigRebuildsExporterOnExportConfigChange(t *testing.T) {
	config := core.DefaultMonitoringConfig()
	config.ExportConfig.Enabled = true
	config.ExportConfig.Port = 9090

	manager, err := NewManager(&config)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	if err = manager.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	oldExporter := manager.GetExporter()
	if oldExporter == nil {
		t.Fatal("expected exporter after initialization")
	}

	newConfig := manager.GetConfig()
	newConfig.ExportConfig.Port = 9091
	if err = manager.UpdateConfig(newConfig); err != nil {
		t.Fatalf("UpdateConfig failed: %v", err)
	}

	if manager.GetExporter() == oldExporter {
		t.Fatal("expected exporter to be rebuilt after export config change")
	}
}
