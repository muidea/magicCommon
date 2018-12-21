package model

import (
	"fmt"
	"reflect"
)

type ptrImpl struct {
	value reflect.Value
}

func (s *ptrImpl) SetValue(val reflect.Value) (err error) {
	if s.value.IsNil() {
		return
	}
	rawVal := reflect.Indirect(s.value)
	if rawVal.Kind() == reflect.Ptr {
		if rawVal.IsNil() {
			return
		}

		rawVal = reflect.Indirect(rawVal)
	}

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return
		}
		val = reflect.Indirect(val)
	}
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return
		}

		val = reflect.Indirect(val)
	}

	if rawVal.Type().String() != val.Type().String() {
		err = fmt.Errorf("can't convert %s to %s", val.Type().String(), rawVal.Type().String())
		return
	}

	rawVal.Set(val)
	return
}

func (s *ptrImpl) GetValue() reflect.Value {
	return s.value
}

func (s *ptrImpl) GetDepend() (ret []reflect.Value, err error) {
	if s.value.IsNil() {
		return
	}

	rawVal := reflect.Indirect(s.value)
	if rawVal.Kind() == reflect.Ptr {
		if rawVal.IsNil() {
			return
		}

		rawVal = reflect.Indirect(rawVal)
	}
	if rawVal.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal ptr, type:%s", rawVal.Type().String())
		return
	}
	if rawVal.Type().String() != "time.Time" {
		ret = append(ret, rawVal)
	}

	return
}

func (s *ptrImpl) GetValueStr() (ret string, err error) {
	if s.value.IsNil() {
		return
	}

	rawVal := reflect.Indirect(s.value)
	if s.value.Kind() == reflect.Ptr {
		if rawVal.IsNil() {
			return
		}
		rawVal = reflect.Indirect(s.value)
	}

	fieldValue, fieldErr := newFieldValue(rawVal)
	if fieldErr != nil {
		err = fieldErr
		return
	}

	ret, err = fieldValue.GetValueStr()
	return
}
