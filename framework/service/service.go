package service

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/framework/plugin/initator"
	"github.com/muidea/magicCommon/framework/plugin/module"
	"github.com/muidea/magicCommon/task"
)

type Service interface {
	Startup(eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) *cd.Result
	Run(block bool)
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
	if errInfo := recover(); errInfo != nil {
		log.Errorf("%s startup failed, err:%+v", s.serviceName, errInfo)
		ret = cd.NewError(cd.UnExpected, "service startup failed")
	}

	initator.Setup(eventHub, backgroundRoutine, nil)
	module.Setup(eventHub, backgroundRoutine, &s.waitGroup)
	log.Infof("%s startup success", s.serviceName)
	s.waitGroup.Wait()
	return
}

func (s *defaultService) Run(block bool) {
	if errInfo := recover(); errInfo != nil {
		log.Errorf("%s run failed, err:%+v", s.serviceName, errInfo)
	}

	initator.Run(nil)
	module.Run(&s.waitGroup)
	log.Infof("%s running!", s.serviceName)
	s.waitGroup.Wait()

	if block {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		sig := <-sigChan
		log.Warnf("%s shutdowning signal:%+v", s.serviceName, sig)
	}
}

func (s *defaultService) Shutdown() {
	if errInfo := recover(); errInfo != nil {
		log.Errorf("%s shutdown failed, err:%+v", s.serviceName, errInfo)
	}

	module.Teardown(&s.waitGroup)
	initator.Teardown(nil)
	log.Infof("%s shutdown success", s.serviceName)
	s.waitGroup.Wait()
}
