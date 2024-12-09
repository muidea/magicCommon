package module

import (
	"fmt"
	"testing"

	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"
)

type Abc interface {
	Hello() bool
}

type abcInfo struct {
}

func (s *abcInfo) Hello() bool {
	fmt.Printf("Abc.Hello")
	return true
}

func (s *abcInfo) Weight() int {
	return 123
}

type Demo struct {
}

func (s *Demo) ID() string {
	fmt.Printf("ID")
	return "abcInfo"
}

func (s *Demo) Weight() int64 {
	return 123
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

func TestWeight(t *testing.T) {
	var demo interface{}
	demo = &Demo{}

	val := weight(demo)
	if val != defaultWeight {
		t.Errorf("check weight faield")
	}

	var abc interface{}
	abc = &abcInfo{}

	val = weight(abc)
	if val != 123 {
		t.Errorf("check weight faield")
	}

}

func TestRegister(t *testing.T) {
	var demo interface{}
	demo = &Demo{}

	Register(demo)
}

func TestSetup(t *testing.T) {
	var demo interface{}
	demo = &Demo{}

	Setup(demo, "abcInfo", nil, nil)
}

func TestTeardown(t *testing.T) {
	var demo interface{}
	demo = &Demo{}

	Teardown(demo)
}

func TestBindRegistry(t *testing.T) {
	var demo interface{}
	demo = &Demo{}

	var a interface{}
	a = &abcInfo{}

	BindRegistry(demo, a)

	BindClient(demo, 100)
}

func TestAppendSlice(t *testing.T) {
	valList := []int{1, 2, 3, 3, 4, 5, 6, 7}
	nv := 10

	nList := []int{}
	if len(valList) == 0 {
		nList = append(nList, nv)
	} else {
		ok := false
		for idx, val := range valList {
			if val <= nv {
				nList = append(nList, val)
				continue
			}

			ok = true
			nList = append(nList, nv)
			nList = append(nList, valList[idx:]...)
			break
		}

		if !ok {
			nList = append(nList, nv)
		}
	}

	t.Log(nList)
}
