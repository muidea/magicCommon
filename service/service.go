package service

import (
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"
)

type Service interface {
	Startup(eventHub event.Hub, backgroundRoutine task.BackgroundRoutine)
	Run()
	Shutdown()
}
