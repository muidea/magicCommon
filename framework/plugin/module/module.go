package module

import (
	"sync"

	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicCommon/framework/plugin/common"
)

var moduleMgr = common.NewPluginMgr("module")

func Register(module interface{}) {
	moduleMgr.Register(module)
}

func Setup(eventHub event.Hub, backgroundRoutine task.BackgroundRoutine, wg *sync.WaitGroup) {
	moduleMgr.Setup(eventHub, backgroundRoutine, wg)
}

func Run(wg *sync.WaitGroup) {
	moduleMgr.Run(wg)
}

func Teardown(wg *sync.WaitGroup) {
	moduleMgr.Teardown(wg)
}
