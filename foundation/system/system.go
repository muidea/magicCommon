package system

import (
	"fmt"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/util"
)

func InvokeEntityFunc(entityVal interface{}, funcName string, params ...interface{}) (err *cd.Result) {
	if entityVal == nil {
		errMsg := "entityVal is nil"
		err = cd.NewResult(cd.IllegalParam, errMsg)
		return
	}

	vVal := reflect.ValueOf(entityVal)
	funcVal := vVal.MethodByName(funcName)
	if !isValidMethod(funcVal) || funcVal.IsZero() {
		errMsg := fmt.Sprintf("no such method:%s", funcName)
		err = cd.NewResult(cd.NoExist, errMsg)
		return
	}

	defer func() {
		if errInfo := recover(); errInfo != nil {
			err = cd.NewResult(cd.UnExpected, fmt.Sprintf("recover! invoke %s unexpected, err:%v\nstack:\n%s", funcName, errInfo, util.GetStack(3)))
		}
	}()

	param, err := prepareParams(funcVal, params)
	if err != nil {
		return err
	}

	rVals := funcVal.Call(param)
	if funcVal.Type().NumOut() == 0 {
		return
	}
	errVal, errOK := rVals[0].Interface().(*cd.Result)
	if !errOK {
		err = cd.NewResult(cd.UnExpected, "invoke method return illegal result")
		return
	}

	err = errVal
	return
}

func isValidMethod(funcVal reflect.Value) bool {
	return funcVal.IsValid() && !funcVal.IsZero()
}

func prepareParams(funcVal reflect.Value, params []interface{}) ([]reflect.Value, *cd.Result) {
	inNum := funcVal.Type().NumIn()
	if inNum == 0 {
		return nil, nil
	}

	paramTypes := make([]reflect.Type, inNum)
	for idx := range params {
		if idx >= inNum {
			break
		}
		paramTypes[idx] = funcVal.Type().In(idx)
	}

	param := make([]reflect.Value, inNum)
	for idx, val := range params {
		if idx >= inNum {
			break
		}

		var err *cd.Result
		param[idx], err = convertParam(val, paramTypes[idx])
		if err != nil {
			return nil, err
		}
	}

	return param, nil
}

func convertParam(val interface{}, expectedType reflect.Type) (reflect.Value, *cd.Result) {
	if val == nil {
		return reflect.New(expectedType).Elem(), nil
	}

	rVal := reflect.ValueOf(val)
	if rVal.Kind() == reflect.Interface {
		rVal = rVal.Elem()
	}

	if !rVal.Type().ConvertibleTo(expectedType) {
		errMsg := fmt.Sprintf("[mismatch param, expect type:%s, value type:%s]", expectedType.String(), rVal.Type().String())
		return reflect.Value{}, cd.NewResult(cd.IllegalParam, errMsg)
	}

	return rVal, nil
}
