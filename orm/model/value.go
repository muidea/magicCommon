package model

import (
	"fmt"
	"reflect"
	"time"
)

// FieldValue FieldValue
type FieldValue interface {
	IsValid() bool
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

func (s *valueImpl) IsValid() bool {
	return s.value.IsValid()
}

func (s *valueImpl) SetValue(val reflect.Value) {
	switch val.Type().Kind() {
	case reflect.Bool:
		switch s.value.Type().Kind() {
		case reflect.Bool:
			s.value.SetBool(val.Bool())
		default:
			msg := fmt.Sprintf("illegal value type, oldType:%v, newType:%v", s.value.Type(), val.Type())
			panic(msg)
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		switch s.value.Type().Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			s.value.SetInt(val.Int())
		default:
			msg := fmt.Sprintf("illegal value type, oldType:%v, newType:%v", s.value.Type(), val.Type())
			panic(msg)
		}
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		switch s.value.Type().Kind() {
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			s.value.SetUint(val.Uint())
		default:
			msg := fmt.Sprintf("illegal value type, oldType:%v, newType:%v", s.value.Type(), val.Type())
			panic(msg)
		}
	case reflect.Float32, reflect.Float64:
		switch s.value.Type().Kind() {
		case reflect.Float32, reflect.Float64:
			s.value.SetFloat(val.Float())
		default:
			msg := fmt.Sprintf("illegal value type, oldType:%v, newType:%v", s.value.Type(), val.Type())
			panic(msg)
		}
	case reflect.String:
		switch s.value.Type().Kind() {
		case reflect.String:
			s.value.SetString(val.String())
		default:
			msg := fmt.Sprintf("illegal value type, oldType:%v, newType:%v", s.value.Type(), val.Type())
			panic(msg)
		}
	case reflect.Struct:
		switch s.value.Type().Kind() {
		case reflect.Struct:
			if val.Type().String() == s.value.Type().String() {
				s.value.Set(val)
			} else {
				msg := fmt.Sprintf("illegal value type, oldType:%v, newType:%v", s.value.Type(), val.Type())
				panic(msg)
			}
		default:
			msg := fmt.Sprintf("illegal value type, oldType:%v, newType:%v", s.value.Type(), val.Type())
			panic(msg)
		}
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
		for idx := 0; idx < pos; {
			sv := val.Slice(idx, idx+1)

			sv = reflect.Indirect(sv)
			ret = append(ret, sv)
			idx++
		}
	}

	return
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
