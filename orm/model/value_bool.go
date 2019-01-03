package model

import (
	"fmt"
	"reflect"
)

type boolImpl struct {
	value reflect.Value
}

func (s *boolImpl) SetValue(val reflect.Value) (err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't set nil ptr")
		return
	}

	rawVal := reflect.Indirect(s.value)
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Bool:
		rawVal.SetBool(val.Bool())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		boolVal := false
		if val.Int() > 0 {
			boolVal = true
		}
		rawVal.Set(reflect.ValueOf(boolVal))
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		boolVal := false
		if val.Uint() != 0 {
			boolVal = true
		}
		rawVal.Set(reflect.ValueOf(boolVal))
	default:
		err = fmt.Errorf("can't convert %s to bool", val.Type().String())
	}
	return
}

func (s *boolImpl) IsNil() bool {
	if s.value.Kind() == reflect.Ptr {
		return s.value.IsNil()
	}

	return false
}

func (s *boolImpl) GetValue() (reflect.Value, error) {
	return s.value, nil
}

func (s *boolImpl) GetDepend() (ret []reflect.Value, err error) {
	// noting todo
	return
}

func (s *boolImpl) GetValueStr() (ret string, err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't get nil ptr value")
		return
	}

	rawVal := reflect.Indirect(s.value)
	if rawVal.Bool() {
		ret = "1"
	} else {
		ret = "0"
	}

	return
}

func (s *boolImpl) Copy() FieldValue {
	return &boolImpl{value: s.value}
}
