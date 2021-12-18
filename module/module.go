package module

import (
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"
)

type Module interface {
	ID() string
	Setup(endpointName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine)
	Teardown()
}

type Service interface {
	Startup(eventHub event.Hub, backgroundRoutine task.BackgroundRoutine)
	Run()
	Shutdown()
}

var moduleList []Module

func Register(module Module) {
	moduleList = append(moduleList, module)
}

func GetModules() []Module {
	return moduleList
}
