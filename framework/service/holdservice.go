package service

import (
	"os"
	"os/signal"
	"syscall"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
)

func HoldService(name string) Service {
	return &holdService{
		defaultService: defaultService{
			serviceName: name,
		},
	}
}

type holdService struct {
	defaultService
}

func (s *holdService) Run() (err *cd.Result) {
	err = s.defaultService.Run()
	if err != nil {
		return
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Warnf("%s shutdowning signal:%+v", s.serviceName, sig)
	return
}
