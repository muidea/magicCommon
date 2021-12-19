package module

import (
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"
	"reflect"
)

type Module interface {
	ID() string
	Setup(endpointName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine)
	Teardown()
}

var moduleList []interface{}

func Register(module interface{}) {
	validModule(module)

	moduleList = append(moduleList, module)
}

func GetModules() []interface{} {
	return moduleList
}

func validModule(ptr interface{}) {
	vType := reflect.TypeOf(ptr)
	if vType.Kind() != reflect.Ptr {
		panic("must be a object ptr")
	}

	vType = vType.Elem()
	_, idOK := vType.FieldByName("ID")
	_, setupOK := vType.FieldByName("Setup")
	_, teardownOK := vType.FieldByName("Teardown")
	if !idOK || !setupOK || !teardownOK {
		panic("invalid module")
	}
}
