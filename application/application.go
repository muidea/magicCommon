package application

import (
	"sync"

	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/module"
	"github.com/muidea/magicCommon/task"
)

type Application interface {
	Startup(service module.Service)
	Run()
	Shutdown()
	EventHub() event.Hub
	BackgroundRoutine() task.BackgroundRoutine
}

var application Application
var applicationOnce sync.Once

func GetApp() Application {
	applicationOnce.Do(func() {
		application = &appImpl{
			backgroundRoutine: task.NewBackgroundRoutine(),
			eventHub:          event.NewHub(),
		}
	})

	return application
}

type appImpl struct {
	backgroundRoutine task.BackgroundRoutine
	eventHub          event.Hub
	service           module.Service
}

func (s *appImpl) Startup(service module.Service) {
	s.service = service
	s.service.Startup()
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
