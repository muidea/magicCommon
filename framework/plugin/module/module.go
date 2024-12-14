package module

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicCommon/framework/plugin/common"
)

var moduleMgr = common.NewPluginMgr("module")

func Register(module interface{}) {
	moduleMgr.Register(module)
}

func Setup(eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) *cd.Result {
	return moduleMgr.Setup(eventHub, backgroundRoutine)
}

func Run() *cd.Result {
	return moduleMgr.Run()
}

func Teardown() {
	moduleMgr.Teardown()
}
