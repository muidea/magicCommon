package application

import (
	"log"
	"sync"

	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/module"
	"github.com/muidea/magicCommon/task"
)

type Application interface {
	EventHub() event.Hub
	BackgroundRoutine() task.BackgroundRoutine
	BindService(service module.Service)
	UnbindService(service module.Service)
	Startup()
	Run()
	Shutdown()
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
	name2Service      sync.Map
}

func (s *appImpl) EventHub() event.Hub {
	return s.eventHub
}

func (s *appImpl) BackgroundRoutine() task.BackgroundRoutine {
	return s.backgroundRoutine
}

func (s *appImpl) BindService(service module.Service) {
	_, ok := s.name2Service.Load(service.Name())
	if ok {
		log.Fatalf("duplicate service, name:%s", service.Name())
		return
	}

	s.name2Service.Store(service.Name(), service)
}

func (s *appImpl) UnbindService(service module.Service) {
	s.name2Service.Delete(service.Name())
}

func (s *appImpl) Startup() {
	var wg sync.WaitGroup
	s.name2Service.Range(func(key, value interface{}) bool {
		service := value.(module.Service)
		wg.Add(1)
		go func() {
			service.Startup()
			wg.Done()
		}()

		return true
	})

	wg.Wait()
}

func (s *appImpl) Run() {
	var wg sync.WaitGroup
	s.name2Service.Range(func(key, value interface{}) bool {
		service := value.(module.Service)
		wg.Add(1)
		go func() {
			service.Run()
			wg.Done()
		}()

		return true
	})

	wg.Wait()
}

func (s *appImpl) Shutdown() {
	var wg sync.WaitGroup
	s.name2Service.Range(func(key, value interface{}) bool {
		service := value.(module.Service)
		wg.Add(1)
		go func() {
			service.Shutdown()
			wg.Done()
		}()

		return true
	})

	wg.Wait()
}
