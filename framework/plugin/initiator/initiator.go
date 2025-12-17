package initiator

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicCommon/framework/plugin/common"
)

var initiatorMgr = common.NewPluginMgr("initiator")

func Register(initiator any) {
	initiatorMgr.Register(initiator)
}

func GetEntity[T any](id string, maskType T) (ret T, err *cd.Error) {
	entityVal, entityErr := initiatorMgr.GetEntity(id)
	if entityErr != nil {
		err = entityErr
		return
	}

	eVal, eOK := entityVal.(T)
	if !eOK {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("initiator:%s type not match", id))
		return
	}

	ret = eVal
	return
}

func Setup(eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) *cd.Error {
	return initiatorMgr.Setup(eventHub, backgroundRoutine)
}

func Run() *cd.Error {
	return initiatorMgr.Run()
}

func Teardown() {
	initiatorMgr.Teardown()
}
