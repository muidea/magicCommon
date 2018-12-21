package model

import (
	"fmt"
	"reflect"
)

type intImpl struct {
	value reflect.Value
}

func (s *intImpl) SetValue(val reflect.Value) (err error) {
	if s.value.Kind() == reflect.Ptr {
		if s.value.IsNil() {
			return
		}
	}

	rawVal := reflect.Indirect(s.value)
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		rawVal.SetInt(val.Int())
	default:
		err = fmt.Errorf("can't convert %s to int", val.Type().String())
	}
	return
}

func (s *intImpl) GetValue() reflect.Value {
	return s.value
}

func (s *intImpl) GetDepend() (ret []reflect.Value, err error) {
	// noting todo
	return
}

func (s *intImpl) GetValueStr() (ret string, err error) {
	if s.value.Kind() == reflect.Ptr {
		if s.value.IsNil() {
			return
		}
	}

	rawVal := reflect.Indirect(s.value)
	ret = fmt.Sprintf("%d", rawVal.Int())

	return
}

type uintImpl struct {
	value reflect.Value
}

func (s *uintImpl) SetValue(val reflect.Value) (err error) {
	if s.value.Kind() == reflect.Ptr {
		if s.value.IsNil() {
			return
		}
	}

	rawVal := reflect.Indirect(s.value)
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		rawVal.SetUint(val.Uint())
	default:
		err = fmt.Errorf("can't convert %s to uint", val.Type().String())
	}
	return
}

func (s *uintImpl) GetValue() reflect.Value {
	return s.value
}

func (s *uintImpl) GetDepend() (ret []reflect.Value, err error) {
	// noting todo
	return
}

func (s *uintImpl) GetValueStr() (ret string, err error) {
	if s.value.Kind() == reflect.Ptr {
		if s.value.IsNil() {
			return
		}
	}

	rawVal := reflect.Indirect(s.value)
	ret = fmt.Sprintf("%d", rawVal.Uint())

	return
}
