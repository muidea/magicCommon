package module

import (
	"fmt"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"
	"testing"
)

type Abc interface {
	Hello() bool
}

type abc struct {
}

func (s *abc) Hello() bool {
	fmt.Printf("Abc.Hello")
	return true
}

type Demo struct {
}

func (s *Demo) ID() string {
	fmt.Printf("ID")
	return "abc"
}

func (s *Demo) Setup(endpointName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) {
	fmt.Printf("Setup, endpointName:%s", endpointName)
}

func (s *Demo) Teardown() {
	fmt.Printf("Teardown")
}

func (s *Demo) BindRegistry(abc Abc) {
	abc.Hello()
}

func TestRegister(t *testing.T) {
	var demo interface{}
	demo = &Demo{}

	Register(demo)
}

func TestSetup(t *testing.T) {
	var demo interface{}
	demo = &Demo{}

	err := Setup(demo, "abc", nil, nil)
	if err != nil {
		t.Errorf("Setup failed, err:%s", err.Error())
	}
}

func TestTeardown(t *testing.T) {
	var demo interface{}
	demo = &Demo{}

	err := Teardown(demo)
	if err != nil {
		t.Errorf("Teardown failed, err:%s", err.Error())
	}
}

func TestBindRegistry(t *testing.T) {
	var demo interface{}
	demo = &Demo{}

	var a interface{}
	a = &abc{}

	err := BindRegistry(demo, a)
	if err != nil {
		t.Errorf("BindRegistry failed, err:%s", err.Error())
	}
}
