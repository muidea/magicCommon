package model

import (
	"fmt"
	"reflect"
)

type structImpl struct {
	value reflect.Value
}

func (s *structImpl) SetValue(val reflect.Value) (err error) {
	if s.value.Kind() == reflect.Ptr {
		if s.value.IsNil() {
			return
		}
	}

	rawVal := reflect.Indirect(s.value)
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Struct:
		if rawVal.Type().String() == val.Type().String() {
			rawVal.Set(val)
		} else {
			err = fmt.Errorf("can't convert %s to %s", val.Type().String(), rawVal.Type().String())
		}
	default:
		err = fmt.Errorf("can't convert %s to %s", val.Type().String(), rawVal.Type().String())
	}
	return
}

func (s *structImpl) GetValue() reflect.Value {
	return s.value
}

func (s *structImpl) GetDepend() (ret []reflect.Value, err error) {
	rawVal := reflect.Indirect(s.value)
	ret = append(ret, rawVal)

	return
}

func (s *structImpl) GetValueStr() (ret string, err error) {
	if s.value.Kind() == reflect.Ptr {
		if s.value.IsNil() {
			return
		}
	}

	rawVal := reflect.Indirect(s.value)
	pkField, pkErr := getStructPrimaryKey(rawVal)
	if pkErr != nil {
		err = pkErr
		return
	}

	ret, err = pkField.GetFieldValue().GetValueStr()
	return
}
