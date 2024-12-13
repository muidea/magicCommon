package module

import (
	"sync"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicCommon/framework/plugin/common"
)

var moduleMgr = common.NewPluginMgr("module")

func Register(module interface{}) {
	moduleMgr.Register(module)
}

func Setup(eventHub event.Hub, backgroundRoutine task.BackgroundRoutine, wg *sync.WaitGroup) *cd.Result {
	return moduleMgr.Setup(eventHub, backgroundRoutine, wg)
}

func Run(wg *sync.WaitGroup) *cd.Result {
	return moduleMgr.Run(wg)
}

func Teardown(wg *sync.WaitGroup) {
	moduleMgr.Teardown(wg)
}
