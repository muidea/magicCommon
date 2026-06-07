package application

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/framework/service"
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

func (m *MockService) Startup(_ context.Context, name string, hub event.Hub, routine task.BackgroundRoutine) *cd.Error {
	m.startupCalled = true
	m.startupName = name
	m.eventHub = hub
	m.backgroundRoutine = routine
	return m.startupError
}

func (m *MockService) Run(_ context.Context) *cd.Error {
	m.runCalled = true
	return m.runError
}

func (m *MockService) Shutdown(_ context.Context) {
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

func resetApp(t *testing.T) {
	t.Helper()
	ResetForTesting()
	t.Cleanup(func() {
		Shutdown(context.Background())
		ResetForTesting()
	})
}

func TestApplicationGet(t *testing.T) {
	resetApp(t)

	// 测试Get函数返回单例
	app1 := Get()
	app2 := Get()

	assert.NotNil(t, app1, "Application should not be nil")
	assert.NotNil(t, app2, "Application should not be nil")
	assert.Equal(t, app1, app2, "Get should return the same instance")
}

func TestApplicationStartupSuccess(t *testing.T) {
	resetApp(t)
	mockService := &MockService{}

	// 测试正常启动
	err := Startup(context.Background(), mockService)

	assert.Nil(t, err, "Startup should succeed without error")
	assert.True(t, mockService.startupCalled, "Service Startup should be called")
	assert.Equal(t, "magicFramework", mockService.startupName, "Default endpoint name should be used")
}

func TestApplicationStartupWithError(t *testing.T) {
	resetApp(t)
	mockService := &MockService{
		startupError: cd.NewError(cd.Unexpected, "startup failed"),
	}

	// 测试启动失败
	err := Startup(context.Background(), mockService)

	assert.NotNil(t, err, "Startup should return error")
	assert.Equal(t, cd.Code(cd.Unexpected), err.Code, "Error code should match")
	assert.True(t, mockService.startupCalled, "Service Startup should be called")
}

func TestApplicationRunSuccess(t *testing.T) {
	resetApp(t)
	mockService := &MockService{}

	// 先启动服务
	_ = Startup(context.Background(), mockService)
	mockService.Reset()

	// 测试运行
	err := Run(context.Background())

	assert.Nil(t, err, "Run should succeed without error")
	assert.True(t, mockService.runCalled, "Service Run should be called")
}

func TestApplicationRunWithoutStartup(t *testing.T) {
	resetApp(t)

	// 获取application实例但不启动服务
	app := Get()

	// 测试未启动直接运行
	err := app.Run(context.Background())

	assert.NotNil(t, err, "Run should return error when service is nil")
	assert.Equal(t, cd.Code(cd.IllegalParam), err.Code, "Error code should be IllegalParam")
}

func TestApplicationRunWithError(t *testing.T) {
	resetApp(t)
	mockService := &MockService{
		runError: cd.NewError(cd.Unexpected, "run failed"),
	}

	// 先启动服务
	_ = Startup(context.Background(), mockService)
	mockService.Reset()
	mockService.runError = cd.NewError(cd.Unexpected, "run failed")

	// 测试运行失败
	err := Run(context.Background())

	assert.NotNil(t, err, "Run should return error")
	assert.Equal(t, cd.Code(cd.Unexpected), err.Code, "Error code should match")
	assert.True(t, mockService.runCalled, "Service Run should be called")
}

func TestApplicationShutdown(t *testing.T) {
	resetApp(t)
	mockService := &MockService{}

	// 先启动服务
	_ = Startup(context.Background(), mockService)
	mockService.Reset()

	// 测试关闭
	Shutdown(context.Background())

	assert.True(t, mockService.shutdownCalled, "Service Shutdown should be called")
}

func TestApplicationShutdownWithoutStartup(t *testing.T) {
	resetApp(t)
	// 测试未启动直接关闭（应该不会panic）
	Shutdown(context.Background())

	// 如果没有panic，测试通过
	assert.True(t, true, "Shutdown should not panic when service is nil")
}

func TestApplicationEventHub(t *testing.T) {
	resetApp(t)
	app := Get()
	hub := app.EventHub()

	assert.NotNil(t, hub, "EventHub should not be nil")
}

func TestApplicationBackgroundRoutine(t *testing.T) {
	resetApp(t)
	app := Get()
	routine := app.BackgroundRoutine()

	assert.NotNil(t, routine, "BackgroundRoutine should not be nil")
}

func TestApplicationInterfaceMethods(t *testing.T) {
	resetApp(t)
	app := Get()
	mockService := &MockService{}

	// 测试接口方法
	err := app.Startup(context.Background(), mockService)
	assert.Nil(t, err, "Startup should succeed")
	assert.True(t, mockService.startupCalled, "Service Startup should be called")

	// 重置并测试Run
	mockService.Reset()
	err = app.Run(context.Background())
	assert.Nil(t, err, "Run should succeed")
	assert.True(t, mockService.runCalled, "Service Run should be called")

	// 测试Shutdown
	mockService.Reset()
	app.Shutdown(context.Background())
	assert.True(t, mockService.shutdownCalled, "Service Shutdown should be called")

	// 测试获取EventHub和BackgroundRoutine
	hub := app.EventHub()
	assert.NotNil(t, hub, "EventHub should not be nil")

	routine := app.BackgroundRoutine()
	assert.NotNil(t, routine, "BackgroundRoutine should not be nil")
}

func TestApplicationRejectsRepeatedStartup(t *testing.T) {
	resetApp(t)
	first := &MockService{}
	second := &MockService{}

	err := Startup(context.Background(), first)
	assert.Nil(t, err)

	err = Startup(context.Background(), second)
	assert.NotNil(t, err)
	assert.Equal(t, cd.Code(cd.IllegalParam), err.Code)
	assert.False(t, second.startupCalled)
}

func TestApplicationAllowsStartupAfterShutdown(t *testing.T) {
	resetApp(t)
	first := &MockService{}
	second := &MockService{}

	assert.Nil(t, Startup(context.Background(), first))
	Shutdown(context.Background())
	assert.Nil(t, Startup(context.Background(), second))
	assert.True(t, second.startupCalled)
}

func TestApplicationStartupWithOptionsUsesExplicitServiceNameAndRuntime(t *testing.T) {
	resetApp(t)
	mockService := &MockService{}
	hub := event.NewHub(2)
	routine := task.NewBackgroundRoutine(2)
	opts := Options{
		ServiceName:       "explicit-service",
		EventHub:          hub,
		BackgroundRoutine: routine,
	}

	err := StartupWithOptions(context.Background(), mockService, opts)
	assert.Nil(t, err)
	assert.Equal(t, "explicit-service", mockService.startupName)
	assert.Equal(t, hub, mockService.eventHub)
	assert.Equal(t, routine, mockService.backgroundRoutine)

	Shutdown(context.Background())
	assert.NoError(t, routine.AsyncFunction(func() {}), "injected routine should not be application-owned by default")
	routine.Shutdown(context.Background())
	hub.Terminate(context.Background())
}

func TestApplicationStartupWithOptionsUsesConfigDir(t *testing.T) {
	resetApp(t)
	tempDir := t.TempDir()
	configContent := `endpointName = "configured-service"`
	err := os.WriteFile(filepath.Join(tempDir, "application.toml"), []byte(configContent), 0644)
	assert.NoError(t, err)

	mockService := &MockService{}
	err = StartupWithOptions(context.Background(), mockService, Options{ConfigDir: tempDir})
	assert.Nil(t, err)
	assert.Equal(t, "configured-service", mockService.startupName)
}

func TestApplicationFailedStartupRequiresShutdownBeforeRetry(t *testing.T) {
	resetApp(t)
	failing := &MockService{startupError: cd.NewError(cd.Unexpected, "startup failed")}
	startupErr := Startup(context.Background(), failing)
	assert.NotNil(t, startupErr)
	assert.True(t, failing.shutdownCalled)

	retry := &MockService{}
	retryErr := Startup(context.Background(), retry)
	assert.NotNil(t, retryErr)
	assert.False(t, retry.startupCalled)

	Shutdown(context.Background())
	assert.Nil(t, Startup(context.Background(), retry))
	assert.True(t, retry.startupCalled)
}

type mockLifecycle struct {
	startupCalled  bool
	runCalled      bool
	shutdownCalled bool
	startupErr     error
	runErr         error
	shutdownErr    error
}

func (m *mockLifecycle) Startup(_ context.Context) error {
	m.startupCalled = true
	return m.startupErr
}

func (m *mockLifecycle) Run(_ context.Context) error {
	m.runCalled = true
	return m.runErr
}

func (m *mockLifecycle) Shutdown(_ context.Context) error {
	m.shutdownCalled = true
	return m.shutdownErr
}

func TestLifecycleAdapterRunsThroughApplication(t *testing.T) {
	resetApp(t)
	lifecycle := &mockLifecycle{}
	adapted := service.AdaptLifecycle("local", lifecycle)

	assert.Nil(t, Startup(context.Background(), adapted))
	assert.True(t, lifecycle.startupCalled)
	assert.Nil(t, Run(context.Background()))
	assert.True(t, lifecycle.runCalled)
	Shutdown(context.Background())
	assert.True(t, lifecycle.shutdownCalled)
}

func TestLifecycleAdapterConvertsErrors(t *testing.T) {
	resetApp(t)
	lifecycle := &mockLifecycle{startupErr: errors.New("boom")}
	err := Startup(context.Background(), service.AdaptLifecycle("local", lifecycle))
	assert.NotNil(t, err)
	assert.Equal(t, cd.Code(cd.Unexpected), err.Code)
}
