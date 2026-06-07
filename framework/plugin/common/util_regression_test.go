package common

import (
	"context"
	"reflect"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"
)

type panicIDPlugin struct{}

func (s *panicIDPlugin) ID() string            { panic("boom") }
func (s *panicIDPlugin) Run(_ context.Context) {}

type panicWeightPlugin struct{}

func (s *panicWeightPlugin) ID() string            { return "panic-weight" }
func (s *panicWeightPlugin) Weight() int           { panic("boom") }
func (s *panicWeightPlugin) Run(_ context.Context) {}

type nilPlugin struct{}

func (s *nilPlugin) ID() string            { return "nil-plugin" }
func (s *nilPlugin) Run(_ context.Context) {}

type duplicatePlugin struct {
	id string
}

func (s *duplicatePlugin) ID() string            { return s.id }
func (s *duplicatePlugin) Run(_ context.Context) {}

func TestPluginMgrRejectsNilPlugin(t *testing.T) {
	pluginMgr := NewPluginMgr("abc")

	if err := pluginMgr.Register(nil); err == nil {
		t.Fatalf("expected nil plugin registration to fail")
	}
}

func TestPluginMgrRejectsTypedNilPlugin(t *testing.T) {
	pluginMgr := NewPluginMgr("abc")
	var pluginPtr *nilPlugin

	if err := pluginMgr.Register(pluginPtr); err == nil {
		t.Fatalf("expected typed nil plugin registration to fail")
	}
}

func TestPluginMgrGetIDReturnsErrorOnPanic(t *testing.T) {
	pluginMgr := NewPluginMgr("abc")

	_, err := pluginMgr.getID(&panicIDPlugin{})
	if err == nil {
		t.Fatalf("expected getID to report panic as error")
	}
}

func TestPluginMgrGetWeightReturnsErrorOnPanic(t *testing.T) {
	pluginMgr := NewPluginMgr("abc")

	weight, err := pluginMgr.getWeight(&panicWeightPlugin{})
	if err == nil {
		t.Fatalf("expected getWeight to report panic as error")
	}
	if weight != DefaultWeight {
		t.Fatalf("expected default weight after panic, got %d", weight)
	}
}

func TestPluginMgrRejectsDuplicatePluginID(t *testing.T) {
	pluginMgr := NewPluginMgr("abc")

	if err := pluginMgr.Register(&duplicatePlugin{id: "dup"}); err != nil {
		t.Fatalf("unexpected first registration error: %v", err)
	}
	if err := pluginMgr.Register(&duplicatePlugin{id: "dup"}); err == nil {
		t.Fatalf("expected duplicate plugin registration to fail")
	}
}

type invalidRunPlugin struct{}

func (s *invalidRunPlugin) ID() string   { return "invalid-run" }
func (s *invalidRunPlugin) Run(_ string) {}

func TestPluginMgrRejectsInvalidRunSignature(t *testing.T) {
	pluginMgr := NewPluginMgr("abc")

	if err := pluginMgr.Register(&invalidRunPlugin{}); err == nil {
		t.Fatalf("expected invalid run signature to fail")
	}
}

type explicitLifecyclePlugin struct {
	id      string
	setup   []string
	run     []string
	tear    []string
	failSet bool
}

func (s *explicitLifecyclePlugin) ID() string {
	return s.id
}

func (s *explicitLifecyclePlugin) Run(_ context.Context) *cd.Error {
	s.run = append(s.run, s.id)
	return nil
}

func (s *explicitLifecyclePlugin) Setup(_ context.Context, _ event.Hub, _ task.BackgroundRoutine) *cd.Error {
	s.setup = append(s.setup, s.id)
	if s.failSet {
		return cd.NewError(cd.Unexpected, "setup failed")
	}
	return nil
}

func (s *explicitLifecyclePlugin) Teardown(_ context.Context) {
	s.tear = append(s.tear, s.id)
}

func TestPluginMgrExplicitInterfaces(t *testing.T) {
	pluginMgr := NewPluginMgr("abc")
	plugin := &explicitLifecyclePlugin{id: "typed"}

	if err := pluginMgr.Register(plugin); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if err := pluginMgr.Setup(context.Background(), nil, nil); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if err := pluginMgr.Run(context.Background()); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	pluginMgr.Teardown(context.Background())

	if !reflect.DeepEqual(plugin.setup, []string{"typed"}) {
		t.Fatalf("unexpected setup calls: %#v", plugin.setup)
	}
	if !reflect.DeepEqual(plugin.run, []string{"typed"}) {
		t.Fatalf("unexpected run calls: %#v", plugin.run)
	}
	if !reflect.DeepEqual(plugin.tear, []string{"typed"}) {
		t.Fatalf("unexpected teardown calls: %#v", plugin.tear)
	}
}

type rollbackPlugin struct {
	id      string
	order   *[]string
	failSet bool
}

func (s *rollbackPlugin) ID() string { return s.id }
func (s *rollbackPlugin) Run(_ context.Context) {
	*s.order = append(*s.order, "run:"+s.id)
}
func (s *rollbackPlugin) Setup(_ context.Context, _ event.Hub, _ task.BackgroundRoutine) *cd.Error {
	*s.order = append(*s.order, "setup:"+s.id)
	if s.failSet {
		return cd.NewError(cd.Unexpected, "setup failed")
	}
	return nil
}
func (s *rollbackPlugin) Teardown(_ context.Context) {
	*s.order = append(*s.order, "teardown:"+s.id)
}

func TestPluginMgrSetupRollbackOnlyCompletedPlugins(t *testing.T) {
	pluginMgr := NewPluginMgr("abc")
	order := []string{}

	if err := pluginMgr.Register(&rollbackPlugin{id: "01", order: &order}); err != nil {
		t.Fatalf("register first failed: %v", err)
	}
	if err := pluginMgr.Register(&rollbackPlugin{id: "02", order: &order}); err != nil {
		t.Fatalf("register second failed: %v", err)
	}
	if err := pluginMgr.Register(&rollbackPlugin{id: "03", order: &order, failSet: true}); err != nil {
		t.Fatalf("register failing failed: %v", err)
	}
	if err := pluginMgr.Register(&rollbackPlugin{id: "04", order: &order}); err != nil {
		t.Fatalf("register fourth failed: %v", err)
	}

	err := pluginMgr.Setup(context.Background(), nil, nil)
	if err == nil {
		t.Fatalf("expected setup failure")
	}

	expected := []string{"setup:01", "setup:02", "setup:03", "teardown:02", "teardown:01"}
	if !reflect.DeepEqual(order, expected) {
		t.Fatalf("unexpected rollback order: %#v", order)
	}
}
