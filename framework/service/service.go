package service

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/framework/plugin/initiator"
	"github.com/muidea/magicCommon/framework/plugin/module"
	"github.com/muidea/magicCommon/task"
	"log/slog"
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
		slog.Error("s.serviceName startup failed, err:%+v", "field", s.serviceName)
		return
	}

	ret = module.Setup(eventHub, backgroundRoutine)
	if ret != nil {
		slog.Error("s.serviceName startup failed, err:%+v", "field", s.serviceName)
		return
	}

	//slog.Info("s.serviceName startup success", "field", s.serviceName)
	return
}

func (s *defaultService) Run() (ret *cd.Error) {
	if errInfo := recover(); errInfo != nil {
		slog.Error("s.serviceName run failed, err:%+v", "field", s.serviceName)
	}

	ret = initiator.Run()
	if ret != nil {
		slog.Error("s.serviceName run failed, err:%+v", "field", s.serviceName)
		return
	}
	ret = module.Run()
	if ret != nil {
		slog.Error("s.serviceName run failed, err:%+v", "field", s.serviceName)
		return
	}

	//slog.Info("s.serviceName running!", "field", s.serviceName)
	return
}

func (s *defaultService) Shutdown() {
	if errInfo := recover(); errInfo != nil {
		slog.Error("s.serviceName shutdown failed, err:%+v", "field", s.serviceName)
	}

	module.Teardown()
	initiator.Teardown()
	//slog.Info("s.serviceName shutdown success", "field", s.serviceName)
}
