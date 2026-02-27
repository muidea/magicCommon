package service

import (
	"os"
	"os/signal"
	"syscall"

	cd "github.com/muidea/magicCommon/def"
	"log/slog"
)

func HoldService() Service {
	return &holdService{
		defaultService: defaultService{},
	}
}

type holdService struct {
	defaultService
}

func (s *holdService) Run() (err *cd.Error) {
	err = s.defaultService.Run()
	if err != nil {
		return
	}

	sigChan := make(chan os.Signal, 1)
	defer close(sigChan)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	slog.Warn("s.serviceName shutdowning signal:%+v", "field", s.serviceName)
	return
}
