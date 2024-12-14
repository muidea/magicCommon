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
	Run() *cd.Result
	Shutdown()
}

func DefaultService(name string) Service {
	return &defaultService{
		serviceName: name,
	}
}

type defaultService struct {
	serviceName string
	waitGroup   sync.WaitGroup
}

func (s *defaultService) Startup(eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) (ret *cd.Result) {
	ret = initator.Setup(eventHub, backgroundRoutine, nil)
	if ret != nil {
		log.Errorf("%s startup failed, err:%+v", s.serviceName, ret)
		return
	}

	ret = module.Setup(eventHub, backgroundRoutine, &s.waitGroup)
	s.waitGroup.Wait()
	if ret != nil {
		log.Errorf("%s startup failed, err:%+v", s.serviceName, ret)
		return
	}

	log.Infof("%s startup success", s.serviceName)
	return
}

func (s *defaultService) Run() (ret *cd.Result) {
	if errInfo := recover(); errInfo != nil {
		log.Errorf("%s run failed, err:%+v", s.serviceName, errInfo)
	}

	ret = initator.Run(nil)
	if ret != nil {
		log.Errorf("%s run failed, err:%+v", s.serviceName, ret)
		return
	}
	ret = module.Run(&s.waitGroup)
	s.waitGroup.Wait()
	if ret != nil {
		log.Errorf("%s run failed, err:%+v", s.serviceName, ret)
		return
	}

	log.Infof("%s running!", s.serviceName)
	return
}

func (s *defaultService) Shutdown() {
	if errInfo := recover(); errInfo != nil {
		log.Errorf("%s shutdown failed, err:%+v", s.serviceName, errInfo)
	}

	module.Teardown(&s.waitGroup)
	initator.Teardown(nil)
	s.waitGroup.Wait()
	log.Infof("%s shutdown success", s.serviceName)
}
