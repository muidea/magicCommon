package model

import (
	"fmt"
	"reflect"
	"time"

	"muidea.com/magicCommon/orm/util"
)

// FieldValue FieldValue
type FieldValue interface {
	SetValue(val reflect.Value) error
	GetValue() reflect.Value
	GetDepend() ([]reflect.Value, error)
	GetValueStr() (string, error)
}

type valueImpl struct {
	value reflect.Value
}

func newFieldValue(val reflect.Value) FieldValue {
	return &valueImpl{value: reflect.Indirect(val)}
}

func (s *valueImpl) setBoolVal(val reflect.Value) (err error) {
	switch s.value.Type().Kind() {
	case reflect.Bool:
		s.value.SetBool(val.Bool())
	case reflect.Ptr:
		if !s.value.IsNil() {
			rawVal := reflect.Indirect(s.value)
			if rawVal.Type().String() == val.Type().String() {
				rawVal.Set(val)
			} else {
				err = fmt.Errorf("illegal value type, oldType:%v, newType:%v", s.value.Type(), val.Type())
			}
		}
	default:
		err = fmt.Errorf("illegal value type, oldType:%v, newType:%v", s.value.Type(), val.Type())
	}

	return
}

func (s *valueImpl) setIntVal(val reflect.Value) (err error) {
	switch s.value.Type().Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		s.value.SetInt(val.Int())
	case reflect.Bool:
		boolVal := false
		if val.Int() > 0 {
			boolVal = true
		}
		s.value.Set(reflect.ValueOf(boolVal))
	case reflect.Ptr:
		if !s.value.IsNil() {
			rawVal := reflect.Indirect(s.value)
			if rawVal.Type().String() == val.Type().String() {
				rawVal.Set(val)
			} else if rawVal.Type().String() == "bool" {
				boolVal := false
				if val.Int() > 0 {
					boolVal = true
				}
				rawVal.Set(reflect.ValueOf(boolVal))
			} else {
				err = fmt.Errorf("illegal value type, oldType:%v, newType:%v", s.value.Type(), val.Type())
			}
		}
	default:
		err = fmt.Errorf("illegal value type, oldType:%v, newType:%v", s.value.Type(), val.Type())
	}

	return
}

func (s *valueImpl) setUintVal(val reflect.Value) (err error) {
	switch s.value.Type().Kind() {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		s.value.SetUint(val.Uint())
	case reflect.Bool:
		boolVal := false
		if val.Uint() != 0 {
			boolVal = true
		}
		s.value.Set(reflect.ValueOf(boolVal))
	case reflect.Ptr:
		if !s.value.IsNil() {
			rawVal := reflect.Indirect(s.value)
			if rawVal.Type().String() == val.Type().String() {
				rawVal.Set(val)
			} else if rawVal.Type().String() == "bool" {
				boolVal := false
				if val.Uint() != 0 {
					boolVal = true
				}
				rawVal.Set(reflect.ValueOf(boolVal))
			} else {
				err = fmt.Errorf("illegal value type, oldType:%v, newType:%v", s.value.Type(), val.Type())
			}
		}
	default:
		err = fmt.Errorf("illegal value type, oldType:%v, newType:%v", s.value.Type(), val.Type())
	}
	return
}

func (s *valueImpl) setFltVal(val reflect.Value) (err error) {
	switch s.value.Type().Kind() {
	case reflect.Float32, reflect.Float64:
		s.value.SetFloat(val.Float())
	case reflect.Ptr:
		if !s.value.IsNil() {
			rawVal := reflect.Indirect(s.value)
			if rawVal.Type().String() == val.Type().String() {
				rawVal.Set(val)
			} else {
				err = fmt.Errorf("illegal value type, oldType:%v, newType:%v", s.value.Type(), val.Type())
			}
		}
	default:
		err = fmt.Errorf("illegal value type, oldType:%v, newType:%v", s.value.Type(), val.Type())
	}

	return
}

func (s *valueImpl) setStrVal(val reflect.Value) (err error) {
	switch s.value.Type().Kind() {
	case reflect.String:
		s.value.SetString(val.String())
	case reflect.Struct:
		if s.value.Type().String() == "time.Time" {
			tmVal, err := time.ParseInLocation("2006-01-02 15:04:05", val.String(), time.Local)
			if err != nil {
				err = fmt.Errorf("illegal value type, oldType:%v, val:%s, err:%s", s.value.Type(), val.String(), err.Error())
			} else {
				s.value.Set(reflect.ValueOf(tmVal))
			}
		} else {
			err = fmt.Errorf("illegal value type, oldType:%v, newType:%v", s.value.Type(), val.Type().String())
		}
	case reflect.Ptr:
		if !s.value.IsNil() {
			rawVal := reflect.Indirect(s.value)
			if rawVal.Type().String() == val.Type().String() {
				rawVal.Set(val)
			} else if rawVal.Type().String() == "time.Time" {
				tmVal, err := time.ParseInLocation("2006-01-02 15:04:05", val.String(), time.Local)
				if err != nil {
					err = fmt.Errorf("illegal value type, oldType:%v, val:%s, err:%s", s.value.Type(), val.String(), err.Error())
				} else {
					rawVal.Set(reflect.ValueOf(tmVal))
				}
			} else {
				err = fmt.Errorf("illegal value type, oldType:%v, newType:%v", s.value.Type(), val.Type())
			}
		}
	default:
		err = fmt.Errorf("illegal value type, oldType:%v, newType:%v", s.value.Type(), val.Type())
	}

	return
}

func (s *valueImpl) setStructVal(val reflect.Value) (err error) {
	switch s.value.Type().Kind() {
	case reflect.Struct:
		if val.Type().String() == s.value.Type().String() {
			s.value.Set(val)
		} else {
			err = fmt.Errorf("illegal value type, oldType:%v, newType:%v", s.value.Type(), val.Type())
		}
	case reflect.Ptr:
		if !s.value.IsNil() {
			rawVal := reflect.Indirect(s.value)
			if rawVal.Type().String() == val.Type().String() {
				rawVal.Set(val)
			} else {
				err = fmt.Errorf("illegal value type, oldType:%v, newType:%v", s.value.Type(), val.Type())
			}
		}
	default:
		err = fmt.Errorf("illegal value type, oldType:%v, newType:%v", s.value.Type(), val.Type())
	}

	return
}

func (s *valueImpl) SetValue(val reflect.Value) (err error) {
	val = reflect.Indirect(val)
	switch val.Type().Kind() {
	case reflect.Bool:
		err = s.setBoolVal(val)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		err = s.setIntVal(val)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		err = s.setUintVal(val)
	case reflect.Float32, reflect.Float64:
		err = s.setFltVal(val)
	case reflect.String:
		err = s.setStrVal(val)
	case reflect.Struct:
		err = s.setStructVal(val)
	default:
		err = fmt.Errorf("no support fileType, %v", val.Kind().String())
	}

	return
}

func (s *valueImpl) GetValue() reflect.Value {
	return s.value
}

func (s *valueImpl) GetDepend() (ret []reflect.Value, err error) {
	val := reflect.Indirect(s.value)
	switch val.Kind() {
	case reflect.Struct:
		if val.Type().String() != "time.Time" {
			ret = append(ret, val)
		}
	case reflect.Slice:
		typeVal, typeErr := getSliceRawTypeValue(val.Type())
		if typeErr != nil {
			err = typeErr
			return
		}
		if typeVal >= util.TypeStructField {
			pos := val.Len()
			for idx := 0; idx < pos; {
				sv := val.Slice(idx, idx+1)

				sv = reflect.Indirect(sv)
				ret = append(ret, sv)
				idx++
			}
		}
	case reflect.Ptr:
		err = fmt.Errorf("no support fileType, %v", val.Type().String())
	}

	return
}

func (s *valueImpl) GetValueStr() (ret string, err error) {
	if s.value.Kind() == reflect.Ptr {
		if s.value.IsNil() {
			return "", fmt.Errorf("can't get nil ptr string value")
		}
	}

	val := reflect.Indirect(s.value)
	switch val.Kind() {
	case reflect.Bool:
		if val.Bool() {
			ret = "1"
		} else {
			ret = "0"
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		ret = fmt.Sprintf("%d", val.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		ret = fmt.Sprintf("%d", val.Uint())
	case reflect.Float32, reflect.Float64:
		ret = fmt.Sprintf("%f", val.Float())
	case reflect.String:
		ret = fmt.Sprintf("'%s'", val.String())
	case reflect.Struct:
		ts, ok := val.Interface().(time.Time)
		if ok {
			ret = fmt.Sprintf("'%s'", ts.Format("2006-01-02 15:04:05"))
		} else {
			err = fmt.Errorf("no support get string value from struct, [%s]", val.Type().String())
		}
	default:
		err = fmt.Errorf("no support get string value from struct, [%s]", val.Type().String())
	}

	return
}
