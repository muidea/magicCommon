package application

import (
	"os"
	"strconv"
	"sync"

	_ "github.com/muidea/magicCommon/foundation/log"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/service"
	"github.com/muidea/magicCommon/task"
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
	Startup(service service.Service) *cd.Result
	Run()
	Shutdown()
	EventHub() event.Hub
	BackgroundRoutine() task.BackgroundRoutine
}

var application Application
var applicationOnce sync.Once

func Startup(service service.Service) *cd.Result {
	app := Get()
	return app.Startup(service)
}

func Run() {
	app := Get()
	app.Run()
}

func Shutdown() {
	app := Get()
	app.Shutdown()
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

type appImpl struct {
	backgroundRoutine task.BackgroundRoutine
	eventHub          event.Hub
	service           service.Service
}

func (s *appImpl) Startup(service service.Service) *cd.Result {
	s.service = service
	return s.service.Startup(s.eventHub, s.backgroundRoutine)
}

func (s *appImpl) Run() {
	if s.service == nil {
		return
	}

	s.service.Run()
}

func (s *appImpl) Shutdown() {
	if s.service == nil {
		return
	}

	s.service.Shutdown()
}

func (s *appImpl) EventHub() event.Hub {
	return s.eventHub
}

func (s *appImpl) BackgroundRoutine() task.BackgroundRoutine {
	return s.backgroundRoutine
}
