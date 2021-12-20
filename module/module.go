package module

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicBatis/client"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"
	"github.com/muidea/magicCommon/toolkit"
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
	vVal := reflect.ValueOf(module)
	setupFun := vVal.MethodByName("Setup")
	if setupFun.IsNil() {
		err = fmt.Errorf("illegal module")
		return
	}

	defer func() {
		if info := recover(); info != nil {
			err = fmt.Errorf("setup unexpect, err:%v", info)
		}
	}()

	param := make([]reflect.Value, 3)
	param[0] = reflect.ValueOf(endpointName)
	if eventHub != nil {
		param[1] = reflect.ValueOf(eventHub)
	} else {
		param[1] = reflect.New(setupFun.Type().In(1)).Elem()
	}
	if backgroundRoutine != nil {
		param[2] = reflect.ValueOf(backgroundRoutine)
	} else {
		param[2] = reflect.New(setupFun.Type().In(2)).Elem()
	}

	setupFun.Call(param)
	return
}

func Teardown(module interface{}) (err error) {
	vVal := reflect.ValueOf(module)
	teardownFun := vVal.MethodByName("Teardown")
	if teardownFun.IsNil() {
		err = fmt.Errorf("illegal module")
		return
	}

	defer func() {
		if info := recover(); info != nil {
			err = fmt.Errorf("teardown unexpect, err:%v", info)
		}
	}()

	param := make([]reflect.Value, 0)
	teardownFun.Call(param)
	return
}

func BindBatisClient(module interface{}, clnt client.Client) (err error) {
	if clnt == nil {
		err = fmt.Errorf("illegal batis client")
		return
	}

	vVal := reflect.ValueOf(module)
	bindFun := vVal.MethodByName("BindBatisClient")
	if bindFun.IsNil() {
		return
	}

	defer func() {
		if info := recover(); info != nil {
			err = fmt.Errorf("bind batis client unexpect, err:%v", info)
		}
	}()

	param := make([]reflect.Value, 1)
	param[0] = reflect.ValueOf(clnt)

	bindFun.Call(param)
	return
}

func BindRegistry(module interface{}, routeRegistry toolkit.RouteRegistry, casRegistry toolkit.CasRegistry, roleRegistry toolkit.RoleRegistry) (err error) {
	if routeRegistry == nil || casRegistry == nil || roleRegistry == nil {
		err = fmt.Errorf("illegal registry")
		return
	}

	vVal := reflect.ValueOf(module)
	bindFun := vVal.MethodByName("BindRegistry")
	if bindFun.IsNil() {
		return
	}

	defer func() {
		if info := recover(); info != nil {
			err = fmt.Errorf("bind registry unexpect, err:%v", info)
		}
	}()

	param := make([]reflect.Value, 3)
	param[0] = reflect.ValueOf(routeRegistry)
	param[1] = reflect.ValueOf(casRegistry)
	param[2] = reflect.ValueOf(roleRegistry)

	bindFun.Call(param)
	return
}
