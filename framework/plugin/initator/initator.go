package initator

import (
	"fmt"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/foundation/system"
	"github.com/muidea/magicCommon/task"

	"github.com/muidea/magicCommon/framework/plugin/common"
)

var initatorList []interface{}

func validInitator(ptr interface{}) {
	vType := reflect.TypeOf(ptr)
	if vType.Kind() != reflect.Ptr {
		panic("must be a pointer")
	}

	_, idOK := vType.MethodByName(common.IdTag)
	_, setupOK := vType.MethodByName(common.SetupTag)
	_, teardownOK := vType.MethodByName(common.TeardownTag)
	if !idOK || !setupOK || !teardownOK {
		panic("invalid initator")
	}
}

func getWeight(ptr interface{}) int {
	vVal := reflect.ValueOf(ptr)
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

func getID(ptr interface{}) string {
	vVal := reflect.ValueOf(ptr)
	funcVal := vVal.MethodByName(common.IdTag)
	defer func() {
		if info := recover(); info != nil {
			err := fmt.Errorf("invoke %s unexpect, %v", common.IdTag, info)
			panic(err)
		}
	}()

	param := make([]reflect.Value, 0)
	values := funcVal.Call(param)
	if len(values) == 0 {
		err := fmt.Errorf("invoke %s unexpect, illegal result", common.IdTag)
		panic(err)
	}

	if funcVal.Type().Out(0).String() != "string" {
		err := fmt.Errorf("invoke %s unexpect, illegal result", common.IdTag)
		panic(err)
	}

	return values[0].String()
}

func Register(initator interface{}) {
	validInitator(initator)

	curWeight := getWeight(initator)
	newList := []interface{}{}
	if len(initatorList) == 0 {
		newList = append(newList, initator)
	} else {
		ok := false
		for idx, val := range initatorList {
			preWeight := getWeight(val)
			if preWeight <= curWeight {
				newList = append(newList, val)
				continue
			}

			ok = true
			newList = append(newList, initator)
			newList = append(newList, initatorList[idx:]...)
			break
		}

		if !ok {
			newList = append(newList, initator)
		}
	}

	initatorList = newList
}

func GetInitatorEntity[T any](id string, _ T) (ret T, err *cd.Result) {
	for _, val := range initatorList {
		idVal := getID(val)
		if idVal == id {
			eVal, eOK := val.(T)
			if !eOK {
				err = cd.NewError(cd.UnExpected, "invalid initator type")
				return
			}

			err = nil
			ret = eVal
			return
		}
	}

	err = cd.NewError(cd.UnExpected, "initator not found")
	return
}

func Setup(endpointName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) {
	for _, val := range initatorList {
		err := system.InvokeEntityFunc(val, common.SetupTag, endpointName, eventHub, backgroundRoutine)
		if err != nil {
			log.Errorf("invoke setup failed, %v", err)
		}
	}
}

func Run() {
	for _, val := range initatorList {
		err := system.InvokeEntityFunc(val, common.RunTag)
		if err != nil {
			log.Errorf("invoke run failed, %v", err)
		}
	}
}

func Teardown() {
	for _, val := range initatorList {
		err := system.InvokeEntityFunc(val, common.TeardownTag)
		if err != nil {
			log.Errorf("invoke teardown failed, %v", err)
		}
	}
}
