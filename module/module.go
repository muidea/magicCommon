package module

import (
	"fmt"
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

func Setup(endpointName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine, module interface{}) (err error) {
	vVal := reflect.ValueOf(module)
	vVal = vVal.Elem()

	setupFun := vVal.MethodByName("Setup")
	if setupFun.IsNil() {
		err = fmt.Errorf("illegal module")
		return
	}

	param := make([]reflect.Value, 3)
	param[0] = reflect.ValueOf(endpointName)
	param[1] = reflect.ValueOf(eventHub)
	param[2] = reflect.ValueOf(backgroundRoutine)

	setupFun.Call(param)
	return
}

func Teardown(module interface{}) (err error) {
	vVal := reflect.ValueOf(module)
	vVal = vVal.Elem()

	teardownFun := vVal.MethodByName("Teardown")
	if teardownFun.IsNil() {
		err = fmt.Errorf("illegal module")
		return
	}

	param := make([]reflect.Value, 0)
	teardownFun.Call(param)
	return
}
