package service

import (
	"fmt"
	"sync"
	"testing"

	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/framework/plugin/initator"
	"github.com/muidea/magicCommon/framework/plugin/module"
	"github.com/muidea/magicCommon/task"
	"github.com/stretchr/testify/assert"
)

type IndexList []int

var sliceLock sync.RWMutex

type MockIndex struct {
	SetupIndexList    *IndexList
	RunIndexList      *IndexList
	TeardownIndexList *IndexList
	Index             int
}

func (s *MockIndex) ID() string {
	return fmt.Sprintf("%02d", s.Index)
}

func (s *MockIndex) Setup(eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) {
	sliceLock.Lock()
	defer sliceLock.Unlock()
	*s.SetupIndexList = append(*s.SetupIndexList, s.Index)
}

func (s *MockIndex) Run() {
	sliceLock.Lock()
	defer sliceLock.Unlock()
	*s.RunIndexList = append(*s.RunIndexList, s.Index)
}

func (s *MockIndex) Teardown() {
	sliceLock.Lock()
	defer sliceLock.Unlock()
	*s.TeardownIndexList = append(*s.TeardownIndexList, s.Index)
}

func TestInitator(t *testing.T) {
	setupList := make(IndexList, 0)
	runList := make(IndexList, 0)
	teardownList := make(IndexList, 0)

	for idx := 0; idx < 10; idx++ {
		p := &MockIndex{
			Index:             idx,
			SetupIndexList:    &setupList,
			RunIndexList:      &runList,
			TeardownIndexList: &teardownList,
		}
		initator.Register(p)
	}

	service := DefaultService("test")
	service.Startup(nil, nil)
	service.Run(false)
	service.Shutdown()

	ok := true
	sVal := 0
	for _, val := range setupList {
		if val >= sVal {
			sVal = val
			continue
		}
		ok = false
		break
	}
	assert.True(t, ok)

	ok = true
	sVal = 0
	for _, val := range runList {
		if val >= sVal {
			sVal = val
			continue
		}
		ok = false
		break
	}
	assert.True(t, ok)

	ok = true
	sVal = 10
	for _, val := range teardownList {
		if val <= sVal {
			sVal = val
			continue
		}
		ok = false
		break
	}
	assert.True(t, ok)

}

func TestModule(t *testing.T) {
	setupList := make(IndexList, 0)
	runList := make(IndexList, 0)
	teardownList := make(IndexList, 0)

	for idx := 0; idx < 10; idx++ {
		p := &MockIndex{
			Index:             idx,
			SetupIndexList:    &setupList,
			RunIndexList:      &runList,
			TeardownIndexList: &teardownList,
		}
		module.Register(p)
	}

	service := DefaultService("test")
	service.Startup(nil, nil)
	service.Run(false)
	service.Shutdown()

	ok := true
	sVal := 0
	for _, val := range setupList {
		if val >= sVal {
			sVal = val
			continue
		}
		ok = false
	}
	assert.True(t, !ok)

	ok = true
	sVal = 0
	for _, val := range runList {
		if val >= sVal {
			sVal = val
			continue
		}
		ok = false
	}
	assert.True(t, !ok)

	ok = true
	sVal = 10
	for _, val := range teardownList {
		if val <= sVal {
			sVal = val
			continue
		}
		ok = false
	}
	assert.True(t, !ok)

}
