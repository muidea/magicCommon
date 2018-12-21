package model

import (
	"fmt"
	"reflect"
)

type floatImpl struct {
	value reflect.Value
}

func (s *floatImpl) SetValue(val reflect.Value) (err error) {
	if s.value.Kind() == reflect.Ptr {
		if s.value.IsNil() {
			return
		}
	}

	rawVal := reflect.Indirect(s.value)
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Float32, reflect.Float64:
		rawVal.SetFloat(val.Float())
	default:
		err = fmt.Errorf("can't convert %s to float", val.Type().String())
	}
	return
}

func (s *floatImpl) GetValue() reflect.Value {
	return s.value
}

func (s *floatImpl) GetDepend() (ret []reflect.Value, err error) {
	// noting todo
	return
}

func (s *floatImpl) GetValueStr() (ret string, err error) {
	if s.value.Kind() == reflect.Ptr {
		if s.value.IsNil() {
			return
		}
	}

	rawVal := reflect.Indirect(s.value)
	ret = fmt.Sprintf("%f", rawVal.Float())

	return
}
