package model

import (
	"fmt"
	"reflect"
)

type floatImpl struct {
	value reflect.Value
}

func (s *floatImpl) SetValue(val reflect.Value) (err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't set nil ptr")
		return
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

func (s *floatImpl) IsNil() bool {
	if s.value.Kind() == reflect.Ptr {
		return s.value.IsNil()
	}

	return false
}

func (s *floatImpl) GetValue() (reflect.Value, error) {
	return s.value, nil
}

func (s *floatImpl) GetDepend() (ret []reflect.Value, err error) {
	// noting todo
	return
}

func (s *floatImpl) GetValueStr() (ret string, err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't get nil ptr value")
		return
	}

	rawVal := reflect.Indirect(s.value)
	ret = fmt.Sprintf("%f", rawVal.Float())

	return
}

func (s *floatImpl) Copy() FieldValue {
	return &floatImpl{value: s.value}
}
