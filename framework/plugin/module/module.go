package module

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/foundation/system"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicCommon/framework/plugin/common"
)

var moduleList []interface{}

func Register(module interface{}) {
	validModule(module)

	curWeight := weight(module)
	newList := []interface{}{}
	if len(moduleList) == 0 {
		newList = append(newList, module)
	} else {
		ok := false
		for idx, val := range moduleList {
			preWeight := weight(val)
			if preWeight <= curWeight {
				newList = append(newList, val)
				continue
			}

			ok = true
			newList = append(newList, module)
			newList = append(newList, moduleList[idx:]...)
			break
		}

		if !ok {
			newList = append(newList, module)
		}
	}

	moduleList = newList
}

func GetModules() []interface{} {
	return moduleList
}

func validModule(ptr interface{}) {
	vType := reflect.TypeOf(ptr)
	if vType.Kind() != reflect.Ptr {
		panic("must be a pointer")
	}

	_, idOK := vType.MethodByName(common.IdTag)
	_, setupOK := vType.MethodByName(common.SetupTag)
	if !idOK || !setupOK {
		panic("invalid module")
	}
}

func weight(module interface{}) int {
	vVal := reflect.ValueOf(module)
	funcVal := vVal.MethodByName(common.WeightTag)
	if !funcVal.IsValid() {
		return common.DefaultWeight
	}

	defer func() {
		if info := recover(); info != nil {
			err := fmt.Errorf("invoke %s unexpect, %v", common.WeightTag, info)
			panic(err)
		}
	}()

	param := make([]reflect.Value, 0)
	values := funcVal.Call(param)
	if len(values) == 0 {
		return common.DefaultWeight
	}

	if funcVal.Type().Out(0).String() != "int" {
		return common.DefaultWeight
	}

	return int(values[0].Int())
}

func Setup(module interface{}, endpointName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) {
	err := system.InvokeEntityFunc(module, common.SetupTag, endpointName, eventHub, backgroundRoutine)
	if err != nil {
		log.Errorf("Setup failed:%s", err)
	}
}

func Run(module interface{}) {
	err := system.InvokeEntityFunc(module, common.RunTag)
	if err != nil {
		log.Errorf("Run failed:%s", err)
	}
}

func Teardown(module interface{}) {
	err := system.InvokeEntityFunc(module, common.TeardownTag)
	if err != nil {
		log.Errorf("Teardown failed:%s", err)
	}
}
