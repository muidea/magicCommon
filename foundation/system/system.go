package system

import (
	"fmt"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
)

func InvokeEntityFunc(entityVal interface{}, funcName string, params ...interface{}) (err *cd.Result) {
	vVal := reflect.ValueOf(entityVal)
	funcVal := vVal.MethodByName(funcName)
	if !funcVal.IsValid() || funcVal.IsZero() {
		errMsg := fmt.Sprintf("no such method:%s", funcName)
		err = cd.NewError(cd.NoExist, errMsg)
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
