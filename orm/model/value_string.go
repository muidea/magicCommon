package model

import (
	"fmt"
	"reflect"
)

type stringImpl struct {
	value reflect.Value
}

func (s *stringImpl) SetValue(val reflect.Value) (err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't set nil ptr")
		return
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

func (s *stringImpl) IsNil() bool {
	if s.value.Kind() == reflect.Ptr {
		return s.value.IsNil()
	}

	return false
}

func (s *stringImpl) GetValue() (reflect.Value, error) {
	return s.value, nil
}

func (s *stringImpl) GetDepend() (ret []reflect.Value, err error) {
	// noting todo
	return
}

func (s *stringImpl) GetValueStr() (ret string, err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't get nil ptr value")
		return
	}

	rawVal := reflect.Indirect(s.value)
	ret = fmt.Sprintf("'%s'", rawVal.String())

	return
}

func (s *stringImpl) Copy() FieldValue {
	return &stringImpl{value: s.value}
}
