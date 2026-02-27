package application

import (
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"
	"github.com/stretchr/testify/assert"
)

// MockService 用于测试的模拟服务
type MockService struct {
	startupCalled     bool
	runCalled         bool
	shutdownCalled    bool
	startupError      *cd.Error
	runError          *cd.Error
	startupName       string
	eventHub          event.Hub
	backgroundRoutine task.BackgroundRoutine
}

func (m *MockService) Startup(name string, hub event.Hub, routine task.BackgroundRoutine) *cd.Error {
	m.startupCalled = true
	m.startupName = name
	m.eventHub = hub
	m.backgroundRoutine = routine
	return m.startupError
}

func (m *MockService) Run() *cd.Error {
	m.runCalled = true
	return m.runError
}

func (m *MockService) Shutdown() {
	m.shutdownCalled = true
}

func (m *MockService) Reset() {
	m.startupCalled = false
	m.runCalled = false
	m.shutdownCalled = false
	m.startupError = nil
	m.runError = nil
	m.startupName = ""
}

func TestApplicationGet(t *testing.T) {
	ResetForTesting()

	// 测试Get函数返回单例
	app1 := Get()
	app2 := Get()

	assert.NotNil(t, app1, "Application should not be nil")
	assert.NotNil(t, app2, "Application should not be nil")
	assert.Equal(t, app1, app2, "Get should return the same instance")
}

func TestApplicationStartupSuccess(t *testing.T) {
	mockService := &MockService{}

	// 测试正常启动
	err := Startup(mockService)

	assert.Nil(t, err, "Startup should succeed without error")
	assert.True(t, mockService.startupCalled, "Service Startup should be called")
	assert.Equal(t, "magicFramework", mockService.startupName, "Default endpoint name should be used")
}

func TestApplicationStartupWithError(t *testing.T) {
	mockService := &MockService{
		startupError: cd.NewError(cd.Unexpected, "startup failed"),
	}

	// 测试启动失败
	err := Startup(mockService)

	assert.NotNil(t, err, "Startup should return error")
	assert.Equal(t, cd.Code(cd.Unexpected), err.Code, "Error code should match")
	assert.True(t, mockService.startupCalled, "Service Startup should be called")
}

func TestApplicationRunSuccess(t *testing.T) {
	mockService := &MockService{}

	// 先启动服务
	_ = Startup(mockService)
	mockService.Reset()

	// 测试运行
	err := Run()

	assert.Nil(t, err, "Run should succeed without error")
	assert.True(t, mockService.runCalled, "Service Run should be called")
}

func TestApplicationRunWithoutStartup(t *testing.T) {
	ResetForTesting()

	// 获取application实例但不启动服务
	app := Get()

	// 测试未启动直接运行
	err := app.Run()

	assert.NotNil(t, err, "Run should return error when service is nil")
	assert.Equal(t, cd.Code(cd.IllegalParam), err.Code, "Error code should be IllegalParam")
}

func TestApplicationRunWithError(t *testing.T) {
	mockService := &MockService{
		runError: cd.NewError(cd.Unexpected, "run failed"),
	}

	// 先启动服务
	_ = Startup(mockService)
	mockService.Reset()
	mockService.runError = cd.NewError(cd.Unexpected, "run failed")

	// 测试运行失败
	err := Run()

	assert.NotNil(t, err, "Run should return error")
	assert.Equal(t, cd.Code(cd.Unexpected), err.Code, "Error code should match")
	assert.True(t, mockService.runCalled, "Service Run should be called")
}

func TestApplicationShutdown(t *testing.T) {
	mockService := &MockService{}

	// 先启动服务
	_ = Startup(mockService)
	mockService.Reset()

	// 测试关闭
	Shutdown()

	assert.True(t, mockService.shutdownCalled, "Service Shutdown should be called")
}

func TestApplicationShutdownWithoutStartup(t *testing.T) {
	// 测试未启动直接关闭（应该不会panic）
	Shutdown()

	// 如果没有panic，测试通过
	assert.True(t, true, "Shutdown should not panic when service is nil")
}

func TestApplicationEventHub(t *testing.T) {
	app := Get()
	hub := app.EventHub()

	assert.NotNil(t, hub, "EventHub should not be nil")
}

func TestApplicationBackgroundRoutine(t *testing.T) {
	app := Get()
	routine := app.BackgroundRoutine()

	assert.NotNil(t, routine, "BackgroundRoutine should not be nil")
}

func TestApplicationInterfaceMethods(t *testing.T) {
	app := Get()
	mockService := &MockService{}

	// 测试接口方法
	err := app.Startup(mockService)
	assert.Nil(t, err, "Startup should succeed")
	assert.True(t, mockService.startupCalled, "Service Startup should be called")

	// 重置并测试Run
	mockService.Reset()
	err = app.Run()
	assert.Nil(t, err, "Run should succeed")
	assert.True(t, mockService.runCalled, "Service Run should be called")

	// 测试Shutdown
	mockService.Reset()
	app.Shutdown()
	assert.True(t, mockService.shutdownCalled, "Service Shutdown should be called")

	// 测试获取EventHub和BackgroundRoutine
	hub := app.EventHub()
	assert.NotNil(t, hub, "EventHub should not be nil")

	routine := app.BackgroundRoutine()
	assert.NotNil(t, routine, "BackgroundRoutine should not be nil")
}
