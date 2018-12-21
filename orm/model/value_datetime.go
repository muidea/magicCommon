package model

import (
	"fmt"
	"reflect"
	"time"
)

type datetimeImpl struct {
	value reflect.Value
}

func (s *datetimeImpl) SetValue(val reflect.Value) (err error) {
	if s.value.Kind() == reflect.Ptr {
		if s.value.IsNil() {
			return
		}
	}

	rawVal := reflect.Indirect(s.value)
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Struct:
		if val.Type().String() == "time.Time" {
			rawVal.Set(rawVal)
		} else {
			err = fmt.Errorf("can't convert %s to datetime", val.Type().String())
		}
	case reflect.String:
		tmVal, err := time.ParseInLocation("2006-01-02 15:04:05", val.String(), time.Local)
		if err != nil {
			err = fmt.Errorf("illegal value, val:%s, err:%s", val.String(), err.Error())
		} else {
			rawVal.Set(reflect.ValueOf(tmVal))
		}
	default:
		err = fmt.Errorf("can't convert %s to datetime", val.Type().String())
	}
	return
}

func (s *datetimeImpl) GetValue() reflect.Value {
	return s.value
}

func (s *datetimeImpl) GetDepend() (ret []reflect.Value, err error) {
	// noting todo
	return
}

func (s *datetimeImpl) GetValueStr() (ret string, err error) {
	if s.value.Kind() == reflect.Ptr {
		if s.value.IsNil() {
			return
		}
	}

	rawVal := reflect.Indirect(s.value)
	ts, ok := rawVal.Interface().(time.Time)
	if ok {
		ret = fmt.Sprintf("'%s'", ts.Format("2006-01-02 15:04:05"))
	} else {
		err = fmt.Errorf("no support get string value from struct, [%s]", rawVal.Type().String())
	}

	return
}
