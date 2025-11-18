package service

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/framework/plugin/initiator"
	"github.com/muidea/magicCommon/framework/plugin/module"
	"github.com/muidea/magicCommon/task"
)

type Service interface {
	Startup(serviceName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) *cd.Error
	Run() *cd.Error
	Shutdown()
}

func DefaultService() Service {
	return &defaultService{}
}

type defaultService struct {
	serviceName string
}

func (s *defaultService) Startup(serviceName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) (ret *cd.Error) {
	s.serviceName = serviceName

	ret = initiator.Setup(eventHub, backgroundRoutine)
	if ret != nil {
		log.Errorf("%s startup failed, err:%+v", s.serviceName, ret)
		return
	}

	ret = module.Setup(eventHub, backgroundRoutine)
	if ret != nil {
		log.Errorf("%s startup failed, err:%+v", s.serviceName, ret)
		return
	}

	//log.Infof("%s startup success", s.serviceName)
	return
}

func (s *defaultService) Run() (ret *cd.Error) {
	if errInfo := recover(); errInfo != nil {
		log.Errorf("%s run failed, err:%+v", s.serviceName, errInfo)
	}

	ret = initiator.Run()
	if ret != nil {
		log.Errorf("%s run failed, err:%+v", s.serviceName, ret)
		return
	}
	ret = module.Run()
	if ret != nil {
		log.Errorf("%s run failed, err:%+v", s.serviceName, ret)
		return
	}

	//log.Infof("%s running!", s.serviceName)
	return
}

func (s *defaultService) Shutdown() {
	if errInfo := recover(); errInfo != nil {
		log.Errorf("%s shutdown failed, err:%+v", s.serviceName, errInfo)
	}

	module.Teardown()
	initiator.Teardown()
	//log.Infof("%s shutdown success", s.serviceName)
}
