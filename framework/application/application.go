package application

import (
	"context"
	"os"
	"strconv"
	"sync"

	_ "log/slog"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicCommon/framework/configuration"
	"github.com/muidea/magicCommon/framework/health"
	"github.com/muidea/magicCommon/framework/service"
)

var defaultBackTaskQueueSize = 10000
var defaultEventHubQueueSize = 500000

func init() {
	taskQueueSize, taskQueueOK := os.LookupEnv("BG_TASK_QUEUE_SIZE")
	if taskQueueOK {
		iVal, iErr := strconv.Atoi(taskQueueSize)
		if iErr == nil && iVal > 0 {
			defaultBackTaskQueueSize = iVal
		}
	}

	eventQueueSize, eventQueueOK := os.LookupEnv("HUB_EVENT_QUEUE_SIZE")
	if eventQueueOK {
		iVal, iErr := strconv.Atoi(eventQueueSize)
		if iErr == nil && iVal > 0 {
			defaultEventHubQueueSize = iVal
		}
	}
}

type Application interface {
	Startup(ctx context.Context, service service.Service) *cd.Error
	Run(ctx context.Context) *cd.Error
	Shutdown(ctx context.Context)
	EventHub() event.Hub
	BackgroundRoutine() task.BackgroundRoutine
}

var application Application
var applicationOnce sync.Once

func Startup(ctx context.Context, service service.Service) *cd.Error {
	return Get().Startup(ctx, service)
}

func Run(ctx context.Context) *cd.Error {
	return Get().Run(ctx)
}

func Shutdown(ctx context.Context) {
	Get().Shutdown(ctx)
}

func Get() Application {
	applicationOnce.Do(func() {
		application = &appImpl{
			backgroundRoutine: task.NewBackgroundRoutine(defaultBackTaskQueueSize),
			eventHub:          event.NewHub(defaultEventHubQueueSize),
		}
	})

	return application
}

// ResetForTesting resets the application singleton for testing purposes
// This should only be used in tests
func ResetForTesting() {
	application = nil
	applicationOnce = sync.Once{}
	health.ResetDefaultManager()
}

type appImpl struct {
	backgroundRoutine task.BackgroundRoutine
	eventHub          event.Hub
	service           service.Service
}

func (s *appImpl) Startup(ctx context.Context, service service.Service) *cd.Error {
	if ctx == nil {
		ctx = context.Background()
	}
	err := configuration.InitDefaultConfigManager("")
	if err != nil {
		return cd.NewError(cd.Unexpected, err.Error())
	}

	nameVal, nameErr := configuration.GetString("endpointName")
	if nameErr != nil {
		nameVal = "magicFramework"
	}

	s.service = service
	return s.service.Startup(ctx, nameVal, s.eventHub, s.backgroundRoutine)
}

func (s *appImpl) Run(ctx context.Context) *cd.Error {
	if s.service == nil {
		return cd.NewError(cd.IllegalParam, "service is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	return s.service.Run(ctx)
}

func (s *appImpl) Shutdown(ctx context.Context) {
	if s.service == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}

	s.service.Shutdown(ctx)
	if s.backgroundRoutine != nil {
		s.backgroundRoutine.Shutdown(ctx)
	}
	if s.eventHub != nil {
		s.eventHub.Terminate(ctx)
	}
	s.service = nil
	s.backgroundRoutine = task.NewBackgroundRoutine(defaultBackTaskQueueSize)
	s.eventHub = event.NewHub(defaultEventHubQueueSize)
	_ = configuration.CloseConfigManager()
	health.ResetDefaultManager()
}

func (s *appImpl) EventHub() event.Hub {
	return s.eventHub
}

func (s *appImpl) BackgroundRoutine() task.BackgroundRoutine {
	return s.backgroundRoutine
}
