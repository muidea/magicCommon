package service

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"
)

type Service interface {
	Startup(eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) *cd.Result
	Run()
	Shutdown()
}
