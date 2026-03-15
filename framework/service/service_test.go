package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/framework/configuration"
	"github.com/muidea/magicCommon/framework/plugin/initiator"
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

func TestInitiator(t *testing.T) {
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
		initiator.Register(p)
	}

	service := DefaultService()
	_ = service.Startup("test", nil, nil)
	_ = service.Run()
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

	service := DefaultService()
	_ = service.Startup("test", nil, nil)
	_ = service.Run()
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

func TestLoadConfiguredDependencies(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "service_dep_config")
	if err != nil {
		t.Fatalf("mkdir temp failed: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	configContent := `
[serviceDependencies.magicCas]
kind = "required"
target = "http://magiccas:8080"

[serviceDependencies.magicFile]
kind = "optional"
target = "http://magicfile:8080"
`

	configPath := filepath.Join(tempDir, "application.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	if err := configuration.InitDefaultConfigManager(tempDir); err != nil {
		t.Fatalf("init config manager failed: %v", err)
	}
	defer func() { _ = configuration.CloseConfigManager() }()

	dependencies, err := loadConfiguredDependencies()
	if err != nil {
		t.Fatalf("loadConfiguredDependencies failed: %v", err)
	}
	if len(dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(dependencies))
	}
}

func TestDefaultServiceStartupFailsOnRequiredDependency(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "service_dep_startup")
	if err != nil {
		t.Fatalf("mkdir temp failed: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	configContent := fmt.Sprintf(`
[serviceDependencies.magicCas]
kind = "required"
target = "%s"
`, "http://127.0.0.1:1")

	configPath := filepath.Join(tempDir, "application.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	if err := configuration.InitDefaultConfigManager(tempDir); err != nil {
		t.Fatalf("init config manager failed: %v", err)
	}
	defer func() { _ = configuration.CloseConfigManager() }()

	service := DefaultService()
	startupErr := service.Startup("test", nil, nil)
	if startupErr == nil {
		t.Fatalf("expected startup failure when required dependency is not ready")
	}
}
