package common

import "testing"

type panicIDPlugin struct{}

func (s *panicIDPlugin) ID() string { panic("boom") }
func (s *panicIDPlugin) Run()       {}

type panicWeightPlugin struct{}

func (s *panicWeightPlugin) ID() string  { return "panic-weight" }
func (s *panicWeightPlugin) Weight() int { panic("boom") }
func (s *panicWeightPlugin) Run()        {}

type nilPlugin struct{}

func (s *nilPlugin) ID() string { return "nil-plugin" }
func (s *nilPlugin) Run()       {}

type duplicatePlugin struct {
	id string
}

func (s *duplicatePlugin) ID() string { return s.id }
func (s *duplicatePlugin) Run()       {}

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
