package module

import (
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicCommon/framework/plugin/common"
)

var moduleMgr = common.NewPluginMgr("module")

func Register(module interface{}) {
	moduleMgr.Register(module)
}

func Setup(endpointName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) {
	moduleMgr.Setup(endpointName, eventHub, backgroundRoutine)
}

func Run() {
	moduleMgr.Run()
}

func Teardown() {
	moduleMgr.Teardown()
}
