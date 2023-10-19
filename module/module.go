package module

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"
)

var moduleList []interface{}

const (
	idTag           = "ID"
	setupTag        = "Setup"
	teardownTag     = "Teardown"
	runTag          = "Run"
	bindClient      = "BindClient"
	bindRegistryTag = "BindRegistry"
)

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
		panic("must be a pointer")
	}

	_, idOK := vType.MethodByName(idTag)
	_, setupOK := vType.MethodByName(setupTag)
	if !idOK || !setupOK {
		panic("invalid module")
	}
}

func Setup(module interface{}, endpointName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) {
	invokeFunc(module, setupTag, endpointName, eventHub, backgroundRoutine)
	return
}

func Run(module interface{}) {
	invokeFunc(module, runTag)
	return
}

func Teardown(module interface{}) {
	invokeFunc(module, teardownTag)
	return
}

func BindClient(module interface{}, clnt interface{}) {
	if clnt == nil {
		panic("illegal client value")
		return
	}

	invokeFunc(module, bindClient, clnt)
	return
}

func BindRegistry(module interface{}, registry ...interface{}) {
	invokeFunc(module, bindRegistryTag, registry...)
	return
}

func invokeFunc(module interface{}, funcName string, params ...interface{}) {
	vVal := reflect.ValueOf(module)
	funcVal := vVal.MethodByName(funcName)
	if !funcVal.IsValid() {
		return
	}

	defer func() {
		if info := recover(); info != nil {
			err := fmt.Errorf("invoke %s unexpect, %v", funcName, info)
			panic(err)
		}
	}()

	param := make([]reflect.Value, len(params))
	for idx, val := range params {
		fType := funcVal.Type().In(idx)
		if val != nil {
			rVal := reflect.ValueOf(val)
			if rVal.Kind() == reflect.Interface {
				rVal = rVal.Elem()
			}

			if rVal.Type().String() != fType.String() && !rVal.Type().Implements(fType) {
				panic(fmt.Sprintf("[mismatch param, expect type:%s, value type:%s]", fType.String(), rVal.Type().String()))
			}

			param[idx] = rVal
		} else {
			param[idx] = reflect.New(fType).Elem()
		}
	}

	funcVal.Call(param)
	return
}
