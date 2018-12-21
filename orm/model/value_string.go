package model

import (
	"fmt"
	"reflect"
)

type stringImpl struct {
	value reflect.Value
}

func (s *stringImpl) SetValue(val reflect.Value) (err error) {
	if s.value.Kind() == reflect.Ptr {
		if s.value.IsNil() {
			return
		}
	}

	rawVal := reflect.Indirect(s.value)
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.String:
		rawVal.SetString(val.String())
	default:
		err = fmt.Errorf("can't convert %s to string", val.Type().String())
	}
	return
}

func (s *stringImpl) GetValue() reflect.Value {
	return s.value
}

func (s *stringImpl) GetDepend() (ret []reflect.Value, err error) {
	// noting todo
	return
}

func (s *stringImpl) GetValueStr() (ret string, err error) {
	if s.value.Kind() == reflect.Ptr {
		if s.value.IsNil() {
			return
		}
	}

	rawVal := reflect.Indirect(s.value)
	ret = fmt.Sprintf("'%s'", rawVal.String())

	return
}
