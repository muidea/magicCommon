package service

import (
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

func DefaultService() Service {
	return &defaultService{}
}

type defaultService struct {
}

func (s *defaultService) Startup(eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) (ret *cd.Result) {
	if errInfo := recover(); errInfo != nil {
		log.Errorf("service startup failed, err:%+v", errInfo)
		ret = cd.NewError(cd.UnExpected, "service startup failed")
	}

	initator.Setup(eventHub, backgroundRoutine)
	module.Setup(eventHub, backgroundRoutine)
	return
}

func (s *defaultService) Run() {
	if errInfo := recover(); errInfo != nil {
		log.Errorf("service run failed, err:%+v", errInfo)
	}

	initator.Run()
	module.Run()
}

func (s *defaultService) Shutdown() {
	if errInfo := recover(); errInfo != nil {
		log.Errorf("service shutdown failed, err:%+v", errInfo)
	}

	module.Teardown()
	initator.Teardown()
}
