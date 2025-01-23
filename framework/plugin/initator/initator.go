package initator

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicCommon/framework/plugin/common"
)

var initatorMgr = common.NewPluginMgr("initator")

func Register(initator interface{}) {
	initatorMgr.Register(initator)
}

func GetEntity[T any](id string, maskType T) (ret T, err *cd.Result) {
	entityVal, entityErr := initatorMgr.GetEntity(id)
	if entityErr != nil {
		err = entityErr
		return
	}

	eVal, eOK := entityVal.(T)
	if !eOK {
		err = cd.NewResult(cd.UnExpected, "initator type not match")
		return
	}

	ret = eVal
	return
}

func Setup(eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) *cd.Result {
	return initatorMgr.Setup(eventHub, backgroundRoutine)
}

func Run() *cd.Result {
	return initatorMgr.Run()
}

func Teardown() {
	initatorMgr.Teardown()
}
