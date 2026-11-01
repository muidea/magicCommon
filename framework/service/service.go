package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"log/slog"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/framework/configuration"
	"github.com/muidea/magicCommon/framework/health"
	"github.com/muidea/magicCommon/framework/plugin/initiator"
	"github.com/muidea/magicCommon/framework/plugin/module"
	"github.com/muidea/magicCommon/task"
)

type Service interface {
	Startup(ctx context.Context, serviceName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) *cd.Error
	Run(ctx context.Context) *cd.Error
	Shutdown(ctx context.Context)
}

// Quiescer is the checked pre-release barrier used by Application shutdown.
type Quiescer interface {
	Quiesce(context.Context) *cd.Error
}

// CheckedShutdown reports final teardown failures without releasing downstream
// runtime dependencies. Successfully released stages must not run twice.
type CheckedShutdown interface {
	ShutdownChecked(context.Context) *cd.Error
}

func DefaultService() Service {
	return &defaultService{}
}

type defaultService struct {
	serviceName       string
	shutdownMu        sync.Mutex
	quiesced          bool
	tornDown          bool
	initiatorsStarted bool
	modulesStarted    bool
	backgroundRoutine task.BackgroundRoutine
	dependencies      []health.Dependency
	probePolicy       health.ProbePolicy
	monitor           *health.DependencyMonitor
}

func (s *defaultService) Startup(ctx context.Context, serviceName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) (ret *cd.Error) {
	if ctx == nil {
		ctx = context.Background()
	}
	s.serviceName = serviceName
	s.backgroundRoutine = backgroundRoutine
	s.monitor = nil
	s.shutdownMu.Lock()
	s.quiesced, s.tornDown = false, false
	s.initiatorsStarted, s.modulesStarted = false, false
	s.shutdownMu.Unlock()
	manager := health.DefaultManager()
	manager.SetService(serviceName)
	manager.MarkStarting()
	s.probePolicy = health.ProbePolicy{
		Interval: time.Duration(configuration.GetIntWithDefault("health.dependencyProbe.intervalSeconds", 5)) * time.Second,
		Timeout:  time.Duration(configuration.GetIntWithDefault("health.dependencyProbe.timeoutSeconds", 3)) * time.Second,
		MaxAge:   time.Duration(configuration.GetIntWithDefault("health.dependencyProbe.maxAgeSeconds", 15)) * time.Second,
	}
	if err := s.probePolicy.Validate(); err != nil {
		manager.MarkFailed(err)
		return err
	}

	s.initiatorsStarted = true
	ret = initiator.Setup(ctx, eventHub, backgroundRoutine)
	if ret != nil {
		manager.MarkFailed(ret)
		slog.Error("service startup failed", "service", s.serviceName, "stage", "initiator.setup", "error", ret)
		return
	}

	dependencies, depErr := loadConfiguredDependencies()
	if depErr != nil {
		ret = cd.NewError(cd.Unexpected, depErr.Error())
		manager.MarkFailed(ret)
		slog.Error("service startup failed", "service", s.serviceName, "stage", "dependency.config", "error", ret)
		return
	}
	s.dependencies = dependencies
	dependencies, ret = manager.Dependencies(dependencies)
	if ret != nil {
		manager.MarkFailed(ret)
		return ret
	}
	checkCtx, cancel := context.WithTimeout(ctx, s.probePolicy.Timeout)
	ret = manager.CheckDependencies(checkCtx, dependencies)
	cancel()
	if ret == nil && ctx.Err() != nil {
		ret = cd.NewError(cd.Timeout, "service startup context is unavailable")
	}
	if ret != nil {
		manager.MarkFailed(ret)
		slog.Error("service startup failed", "service", s.serviceName, "stage", "dependency.check", "error", ret)
		return
	}

	s.modulesStarted = true
	ret = module.Setup(ctx, eventHub, backgroundRoutine)
	if ret != nil {
		manager.MarkFailed(ret)
		slog.Error("service startup failed", "service", s.serviceName, "stage", "module.setup", "error", ret)
		return
	}

	//slog.Info("s.serviceName startup success", "field", s.serviceName)
	return
}

func (s *defaultService) Run(ctx context.Context) (ret *cd.Error) {
	if ctx == nil {
		ctx = context.Background()
	}
	manager := health.DefaultManager()
	defer func() {
		if errInfo := recover(); errInfo != nil {
			ret = cd.NewError(cd.Unexpected, "service run panicked")
			manager.MarkFailed(ret)
			slog.Error("service run panicked", "service", s.serviceName, "panic", errInfo)
		}
	}()

	ret = initiator.Run(ctx)
	if ret != nil {
		manager.MarkFailed(ret)
		slog.Error("service run failed", "service", s.serviceName, "stage", "initiator.run", "error", ret)
		return
	}
	ret = module.Run(ctx)
	if ret != nil {
		manager.MarkFailed(ret)
		slog.Error("service run failed", "service", s.serviceName, "stage", "module.run", "error", ret)
		return
	}

	dependencies, depErr := manager.Dependencies(s.dependencies)
	if depErr != nil {
		manager.MarkFailed(depErr)
		return depErr
	}
	s.monitor, ret = manager.StartDependencyMonitoring(ctx, s.backgroundRoutine, dependencies, s.probePolicy)
	if ret != nil {
		manager.MarkFailed(ret)
		return ret
	}
	manager.MarkReady()

	//slog.Info("s.serviceName running!", "field", s.serviceName)
	return
}

func loadConfiguredDependencies() ([]health.Dependency, error) {
	configManager := configuration.GetDefaultConfigManager()
	if configManager == nil {
		return nil, nil
	}

	exported, err := configuration.ExportAllConfigs()
	if err != nil {
		return nil, err
	}

	applicationCfg, ok := exported["application"].(map[string]any)
	if !ok {
		return nil, nil
	}

	dependenciesValue, exists := applicationCfg["serviceDependencies"]
	if !exists {
		return nil, nil
	}

	dependenciesMap, ok := dependenciesValue.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("serviceDependencies must be a map")
	}

	jsonBytes, err := json.Marshal(dependenciesMap)
	if err != nil {
		return nil, err
	}

	decoded := map[string]health.Dependency{}
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		return nil, err
	}

	ret := make([]health.Dependency, 0, len(decoded))
	for name, dep := range decoded {
		if dep.Kind == "" {
			dep.Kind = health.RequiredDependency
		}
		dep.Name = name
		ret = append(ret, dep)
	}

	return ret, nil
}

func (s *defaultService) Shutdown(ctx context.Context) {
	if err := s.ShutdownChecked(ctx); err != nil {
		slog.Error("service shutdown incomplete; dependencies retained", "error", err)
	}
}

func (s *defaultService) ShutdownChecked(ctx context.Context) *cd.Error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := s.Quiesce(ctx); err != nil {
		return err
	}
	s.shutdownMu.Lock()
	defer s.shutdownMu.Unlock()
	if s.tornDown {
		return nil
	}
	if s.modulesStarted {
		if err := module.TeardownChecked(ctx); err != nil {
			return err
		}
		s.modulesStarted = false
	}
	if s.initiatorsStarted {
		if err := initiator.TeardownChecked(ctx); err != nil {
			return err
		}
		s.initiatorsStarted = false
	}
	s.tornDown = true
	return nil
}

func (s *defaultService) Quiesce(ctx context.Context) *cd.Error {
	if ctx == nil {
		ctx = context.Background()
	}
	s.shutdownMu.Lock()
	defer s.shutdownMu.Unlock()
	if s.quiesced {
		return nil
	}
	health.DefaultManager().MarkStopping()
	s.monitor.BeginShutdown()
	// Notify every input owner before waiting for any work to drain, even if
	// another owner's notification fails or a dependency probe ignores cancel.
	var initErr, moduleErr *cd.Error
	if s.initiatorsStarted {
		initErr = initiator.BeginShutdown(ctx)
	}
	if s.modulesStarted {
		moduleErr = module.BeginShutdown(ctx)
	}
	if initErr != nil {
		return initErr
	}
	if moduleErr != nil {
		return moduleErr
	}
	if err := s.monitor.Stop(ctx); err != nil {
		return err
	}
	if s.initiatorsStarted {
		if err := initiator.Quiesce(ctx); err != nil {
			return err
		}
	}
	if s.modulesStarted {
		if err := module.Quiesce(ctx); err != nil {
			return err
		}
	}
	s.quiesced = true
	return nil
}
