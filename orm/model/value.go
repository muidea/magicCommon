package model

import (
	"fmt"
	"reflect"
)

// FieldValue FieldValue
type FieldValue interface {
	SetValue(val reflect.Value) error
	IsNil() bool
	GetValue() (reflect.Value, error)
	GetDepend() ([]reflect.Value, error)
	GetValueStr() (string, error)
}

func newFieldValue(val reflect.Value) (ret FieldValue, err error) {
	rawVal := reflect.Indirect(val)
	switch rawVal.Kind() {
	case reflect.Bool:
		ret = &boolImpl{value: val}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		ret = &intImpl{value: val}
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		ret = &uintImpl{value: val}
	case reflect.Float32, reflect.Float64:
		ret = &floatImpl{value: val}
	case reflect.String:
		ret = &stringImpl{value: val}
	case reflect.Struct:
		if rawVal.Type().String() == "time.Time" {
			ret = &datetimeImpl{value: val}
		} else {
			ret = &structImpl{value: val}
		}
	case reflect.Slice:
		ret = &sliceImpl{value: val}
	case reflect.Ptr:
		if rawVal.IsNil() {
			ret = &nilImpl{value: val}
			return
		}

		rawVal = reflect.Indirect(rawVal)
		val = reflect.Indirect(val)
		switch rawVal.Kind() {
		case reflect.Bool:
			ret = &boolImpl{value: val}
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			ret = &intImpl{value: val}
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			ret = &uintImpl{value: val}
		case reflect.Float32, reflect.Float64:
			ret = &floatImpl{value: val}
		case reflect.String:
			ret = &stringImpl{value: val}
		case reflect.Struct:
			if rawVal.Type().String() == "time.Time" {
				ret = &datetimeImpl{value: val}
			} else {
				ret = &structImpl{value: val}
			}
		case reflect.Slice:
			ret = &sliceImpl{value: val}
		default:
			err = fmt.Errorf("no support value ptr type, type:%s", val.Type().String())
		}
	default:
		err = fmt.Errorf("no support value type, kind:%s, type:%s", val.Kind().String(), val.Type().String())
	}

	return
}
