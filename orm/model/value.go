package model

import (
	"fmt"
	"reflect"
	"time"
)

// FieldValue FieldValue
type FieldValue interface {
	IsPtr() bool
	SetValue(val reflect.Value)
	GetValue() reflect.Value
	GetDepend() []reflect.Value
	GetValueStr() string
}

type valueImpl struct {
	value reflect.Value
}

func newFieldValue(val reflect.Value) FieldValue {
	return &valueImpl{value: val}
}

func (s *valueImpl) IsPtr() bool {
	return s.value.Type().Kind() == reflect.Ptr
}

func (s *valueImpl) SetValue(val reflect.Value) {
	switch val.Type().Kind() {
	case reflect.Bool:
		s.value.SetBool(val.Bool())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		s.value.SetInt(val.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		s.value.SetUint(val.Uint())
	case reflect.Float32, reflect.Float64:
		s.value.SetFloat(val.Float())
	case reflect.String:
		s.value.SetString(val.String())
	case reflect.Struct:
		s.value.Set(val)
	default:
		msg := fmt.Sprintf("no support fileType, %v", val.Kind())
		panic(msg)
	}
}

func (s *valueImpl) GetValue() reflect.Value {
	return s.value
}

func (s *valueImpl) GetDepend() (ret []reflect.Value) {
	val := reflect.Indirect(s.value)
	switch val.Kind() {
	case reflect.Struct:
		if val.Type().String() != "time.Time" {
			ret = append(ret, s.value)
		}
	case reflect.Slice:
		pos := val.Len()
		for idx := 0; idx < pos; idx++ {
			sv := val.Slice(idx, idx+1)

			sv = reflect.Indirect(sv)
			ret = append(ret, sv)
		}
	}

	return nil
}

func (s *valueImpl) GetValueStr() (ret string) {
	val := reflect.Indirect(s.value)
	switch val.Kind() {
	case reflect.Bool:
		if s.value.Bool() {
			ret = "1"
		} else {
			ret = "0"
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		ret = fmt.Sprintf("%d", s.value.Interface())
	case reflect.Float32, reflect.Float64:
		ret = fmt.Sprintf("%f", s.value.Interface())
	case reflect.String:
		ret = fmt.Sprintf("'%s'", s.value.Interface())
	case reflect.Struct:
		ts, ok := s.value.Interface().(time.Time)
		if ok {
			ret = fmt.Sprintf("'%s'", ts.Format("2006-01-02 15:04:05"))
		} else {
			msg := fmt.Sprintf("illegal value,[%v]", s.value.Interface())
			panic(msg)
		}
	default:
		msg := fmt.Sprintf("no support fileType, %v", val.Kind())
		panic(msg)
	}

	return
}
