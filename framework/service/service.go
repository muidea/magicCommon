package service

import (
	"sync"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/framework/plugin/initator"
	"github.com/muidea/magicCommon/framework/plugin/module"
	"github.com/muidea/magicCommon/task"
)

type Service interface {
	Startup(eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) *cd.Result
	Run()
	Shutdown()
}

func DefaultService(name string) Service {
	return &defaultService{
		serviceName: name,
	}
}

type defaultService struct {
	serviceName string
}

func (s *defaultService) Startup(eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) (ret *cd.Result) {
	if errInfo := recover(); errInfo != nil {
		log.Errorf("%s startup failed, err:%+v", s.serviceName, errInfo)
		ret = cd.NewError(cd.UnExpected, "service startup failed")
	}

	wg := sync.WaitGroup{}
	initator.Setup(eventHub, backgroundRoutine, nil)
	module.Setup(eventHub, backgroundRoutine, &wg)
	log.Infof("%s startup success", s.serviceName)
	return
}

func (s *defaultService) Run() {
	if errInfo := recover(); errInfo != nil {
		log.Errorf("%s run failed, err:%+v", s.serviceName, errInfo)
	}

	wg := sync.WaitGroup{}
	initator.Run(nil)
	module.Run(&wg)
	log.Infof("%s run success", s.serviceName)
}

func (s *defaultService) Shutdown() {
	if errInfo := recover(); errInfo != nil {
		log.Errorf("%s shutdown failed, err:%+v", s.serviceName, errInfo)
	}

	wg := sync.WaitGroup{}
	module.Teardown(&wg)
	initator.Teardown(nil)
	log.Infof("%s shutdown success", s.serviceName)
}
