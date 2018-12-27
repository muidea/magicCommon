package model

import (
	"fmt"
	"reflect"
)

type nilImpl struct {
	value      reflect.Value
	fieldValue FieldValue
}

func (s *nilImpl) SetValue(val reflect.Value) (err error) {
	if val.Kind() != reflect.Ptr {
		err = fmt.Errorf("can't convert %s to %s", val.Type().String(), s.value.Type().String())
		return
	}

	indirectVal := reflect.Indirect(s.value)
	if indirectVal.Type().String() == val.Type().String() {
		indirectVal.Set(val)
	}

	fieldValue, fieldErr := newFieldValue(val)
	if fieldErr != nil {
		err = fieldErr
		return
	}
	s.fieldValue = fieldValue

	return
}

func (s *nilImpl) IsNil() bool {
	return s.fieldValue == nil
}

func (s *nilImpl) GetValue() (ret reflect.Value, err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't get nil value")
		return
	}

	return s.value, nil
}

func (s *nilImpl) GetDepend() (ret []reflect.Value, err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't get nil depend")
		return
	}

	return s.fieldValue.GetDepend()
}

func (s *nilImpl) GetValueStr() (ret string, err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't get nil ptr value string")
		return
	}

	return s.fieldValue.GetValueStr()
}
