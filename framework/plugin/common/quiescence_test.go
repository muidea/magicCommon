package common

import (
	"context"
	cd "github.com/muidea/magicCommon/def"
	"reflect"
	"testing"
)

type shutdownOwner struct {
	name   string
	calls  *[]string
	panics bool
	reject bool
}

func (s *shutdownOwner) ID() string                    { return s.name }
func (s *shutdownOwner) Run(context.Context) *cd.Error { return nil }
func (s *shutdownOwner) BeginShutdown(context.Context) {
	*s.calls = append(*s.calls, s.name+":close")
	if s.panics {
		panic("cancel failed")
	}
}
func (s *shutdownOwner) Quiesce(context.Context) *cd.Error {
	*s.calls = append(*s.calls, s.name+":drain")
	if s.reject {
		return cd.NewError(cd.Timeout, "busy")
	}
	return nil
}
func (s *shutdownOwner) Teardown(context.Context) { *s.calls = append(*s.calls, s.name+":release") }

func TestShutdownNotifiesAllOwnersBeforeDrainAndRetainsOnFailure(t *testing.T) {
	manager := NewPluginMgr("guard-test")
	var calls []string
	a, b := &shutdownOwner{name: "a", calls: &calls}, &shutdownOwner{name: "b", calls: &calls, reject: true}
	if err := manager.Register(a); err != nil {
		t.Fatal(err)
	}
	if err := manager.Register(b); err != nil {
		t.Fatal(err)
	}
	if err := manager.BeginShutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := manager.Quiesce(context.Background()); err == nil {
		t.Fatal("busy owner accepted")
	}
	if !reflect.DeepEqual(calls, []string{"a:close", "b:close", "a:drain", "b:drain"}) {
		t.Fatal(calls)
	}
	calls = nil
	a.panics = true
	if err := manager.BeginShutdown(context.Background()); err == nil {
		t.Fatal("panic discarded")
	}
	if !reflect.DeepEqual(calls, []string{"a:close", "b:close"}) {
		t.Fatal("one panic prevented other owners closing", calls)
	}
}
