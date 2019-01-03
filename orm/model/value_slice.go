package model

import (
	"encoding/json"
	"fmt"
	"reflect"

	"muidea.com/magicCommon/orm/util"
)

type sliceImpl struct {
	value reflect.Value
}

func (s *sliceImpl) SetValue(val reflect.Value) (err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't set nil ptr")
		return
	}

	rawVal := reflect.Indirect(s.value)
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Slice:
		if val.Type().String() == rawVal.Type().String() {
			rawVal.Set(val)
		} else {
			err = fmt.Errorf("can't convert %s to %s", val.Type().String(), rawVal.Type().String())
		}
	default:
		err = fmt.Errorf("can't convert %s to %s", val.Type().String(), rawVal.Type().String())
	}
	return
}

func (s *sliceImpl) IsNil() bool {
	if s.value.Kind() == reflect.Ptr {
		return s.value.IsNil()
	}

	return false
}

func (s *sliceImpl) GetValue() (reflect.Value, error) {
	return s.value, nil
}

func (s *sliceImpl) GetDepend() (ret []reflect.Value, err error) {
	if s.value.Kind() == reflect.Ptr {
		if s.value.IsNil() {
			return
		}
	}

	rawVal := reflect.Indirect(s.value)
	sliceRawTypeEnum, sliceTypeErr := util.GetSliceRawTypeEnum(rawVal.Type())
	if sliceTypeErr != nil {
		err = sliceTypeErr
		return
	}
	if util.IsBasicType(sliceRawTypeEnum) {
		return
	}

	if !util.IsStructType(sliceRawTypeEnum) {
		err = fmt.Errorf("no support slice element type")
		return
	}

	pos := rawVal.Len()
	for idx := 0; idx < pos; {
		sv := rawVal.Slice(idx, idx+1)

		ret = append(ret, sv)
		idx++
	}
	return
}

func (s *sliceImpl) GetValueStr() (ret string, err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't get nil ptr value")
		return
	}

	rawVal := reflect.Indirect(s.value)
	sliceRawTypeEnum, sliceTypeErr := util.GetSliceRawTypeEnum(rawVal.Type())
	if sliceTypeErr != nil {
		err = sliceTypeErr
		return
	}
	if !util.IsBasicType(sliceRawTypeEnum) {
		err = fmt.Errorf("no support get value string, type:%s", rawVal.Type().String())
		return
	}

	valSlice := []interface{}{}
	pos := rawVal.Len()
	for idx := 0; idx < pos; {
		sv := rawVal.Index(idx)

		sv = reflect.Indirect(sv)
		if sv.Kind() == reflect.Struct {
			datetimeVal := &datetimeImpl{value: sv}
			datetimeStr, _ := datetimeVal.GetValueStr()
			valSlice = append(valSlice, datetimeStr)
		} else {
			valSlice = append(valSlice, sv.Interface())
		}
		idx++
	}

	data, dataErr := json.Marshal(valSlice)
	if dataErr != nil {
		err = dataErr
	}
	ret = string(data)

	return
}

func (s *sliceImpl) Copy() FieldValue {
	return &sliceImpl{value: s.value}
}
