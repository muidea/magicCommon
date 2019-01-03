package model

import (
	"fmt"
	"reflect"
)

type intImpl struct {
	value reflect.Value
}

func (s *intImpl) SetValue(val reflect.Value) (err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't set nil ptr")
		return
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

func (s *intImpl) IsNil() bool {
	if s.value.Kind() == reflect.Ptr {
		return s.value.IsNil()
	}

	return false
}

func (s *intImpl) GetValue() (reflect.Value, error) {
	return s.value, nil
}

func (s *intImpl) GetDepend() (ret []reflect.Value, err error) {
	// noting todo
	return
}

func (s *intImpl) GetValueStr() (ret string, err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't get nil ptr value")
		return
	}

	rawVal := reflect.Indirect(s.value)
	ret = fmt.Sprintf("%d", rawVal.Int())

	return
}

func (s *intImpl) Copy() FieldValue {
	return &intImpl{value: s.value}
}

type uintImpl struct {
	value reflect.Value
}

func (s *uintImpl) SetValue(val reflect.Value) (err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't set nil ptr")
		return
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

func (s *uintImpl) IsNil() bool {
	if s.value.Kind() == reflect.Ptr {
		return s.value.IsNil()
	}

	return false
}

func (s *uintImpl) GetValue() (reflect.Value, error) {
	return s.value, nil
}

func (s *uintImpl) GetDepend() (ret []reflect.Value, err error) {
	// noting todo
	return
}

func (s *uintImpl) GetValueStr() (ret string, err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't get nil ptr value")
		return
	}

	rawVal := reflect.Indirect(s.value)
	ret = fmt.Sprintf("%d", rawVal.Uint())

	return
}

func (s *uintImpl) Copy() FieldValue {
	return &uintImpl{value: s.value}
}
