package module

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"
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

	_, idOK := vType.MethodByName("ID")
	_, setupOK := vType.MethodByName("Setup")
	_, teardownOK := vType.MethodByName("Teardown")
	if !idOK || !setupOK || !teardownOK {
		panic("invalid module")
	}
}

func Setup(module interface{}, endpointName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) (err error) {
	err = invokeFunc(module, "Setup", endpointName, eventHub, backgroundRoutine)
	return
}

func Teardown(module interface{}) (err error) {
	err = invokeFunc(module, "Teardown")
	return
}

func BindBatisClient(module interface{}, clnt interface{}) (err error) {
	if clnt == nil {
		err = fmt.Errorf("illegal batis client")
		return
	}

	err = invokeFunc(module, "BindBatisClient", clnt)
	return
}

func BindRegistry(module interface{}, registry ...interface{}) (err error) {
	err = invokeFunc(module, "BindRegistry", registry...)
	return
}

func invokeFunc(module interface{}, funcName string, params ...interface{}) (err error) {
	vVal := reflect.ValueOf(module)
	funcVal := vVal.MethodByName(funcName)
	if funcVal.IsNil() {
		return
	}

	defer func() {
		if info := recover(); info != nil {
			err = fmt.Errorf("invoke %s unexpect, err:%v", funcName, info)
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

			if fType.String() != rVal.Type().String() {
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
