package initator

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicCommon/framework/plugin/common"
)

var initatorMgr = common.NewPluginMgr("initator")

func Register(initator interface{}) {
	initatorMgr.Register(initator)
}

func GetEntity[T any](id string, maskType T) (ret T, err *cd.Error) {
	entityVal, entityErr := initatorMgr.GetEntity(id)
	if entityErr != nil {
		err = entityErr
		return
	}

	eVal, eOK := entityVal.(T)
	if !eOK {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("initator:%s type not match", id))
		return
	}

	ret = eVal
	return
}

func Setup(eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) *cd.Error {
	return initatorMgr.Setup(eventHub, backgroundRoutine)
}

func Run() *cd.Error {
	return initatorMgr.Run()
}

func Teardown() {
	initatorMgr.Teardown()
}
