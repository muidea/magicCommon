package util

import (
	"fmt"
	"reflect"
	"time"
)

// Define the Type enum
const (
	TypeBooleanField = 1 << iota
	TypeStringField
	TypeDateTimeField
	TypeBitField
	TypeSmallIntegerField
	TypeInteger32Field
	TypeIntegerField
	TypeBigIntegerField
	TypePositiveBitField
	TypePositiveSmallIntegerField
	TypePositiveInteger32Field
	TypePositiveIntegerField
	TypePositiveBigIntegerField
	TypeFloatField
	TypeDoubleField
	TypeStructField
	TypeSliceField
	TypePtrField
)

// Field type
const (
	TypeBaseTypeField = iota
	TypeReferenceField
	TypeReferencePtrField
)

// IsSimpleField IsSimpleField
func IsSimpleField(typeValue int) bool {
	return typeValue < TypeStructField
}

// GetInitValue GetInitValue
func GetInitValue(typeValue int) (ret interface{}) {
	switch typeValue {
	case TypeBooleanField,
		TypeBitField, TypeSmallIntegerField, TypeIntegerField, TypeInteger32Field, TypeBigIntegerField:
		val := 0
		ret = &val
		break
	case TypePositiveBitField, TypePositiveSmallIntegerField, TypePositiveIntegerField, TypePositiveInteger32Field, TypePositiveBigIntegerField:
		val := uint(0)
		ret = &val
		break
	case TypeStringField, TypeDateTimeField:
		val := ""
		ret = &val
		break
	case TypeFloatField, TypeDoubleField:
		val := 0.00
		ret = &val
		break
	default:
		msg := fmt.Sprintf("no support fileType, %d", typeValue)
		panic(msg)
	}
	return
}

// GetValue get value by special type
func GetValue(typeValue int, val reflect.Value) (ret reflect.Value) {
	switch typeValue {
	case TypeBooleanField:
		switch val.Type().Kind() {
		case reflect.Int:
			v := false
			if val.Int() > 0 {
				v = true
			}
			ret = reflect.ValueOf(v)
		case reflect.Bool:
			ret = val
		default:
			msg := fmt.Sprintf("unexception value, exception type:[%d], current value:[%v]", typeValue, val.Interface())
			panic(msg)
		}
	case TypeDateTimeField:
		switch val.Type().Kind() {
		case reflect.String:
			ts, err := time.ParseInLocation("2006-01-02 15:04:05", val.String(), time.Local)
			if err != nil {
				msg := fmt.Sprintf("unexception value, exception type:[%d], current value:[%v]", typeValue, val.Interface())
				panic(msg)
			}
			ret = reflect.ValueOf(ts)
		case reflect.Struct:
			if val.Type().String() == "time.Time" {
				ret = val
			} else {
				msg := fmt.Sprintf("unexception value, exception type:[%d], current value:[%v]", typeValue, val.Interface())
				panic(msg)
			}
		default:
			msg := fmt.Sprintf("unexception value, exception type:[%d], current value:[%v]", typeValue, val.Interface())
			panic(msg)
		}
	case TypeBitField, TypeSmallIntegerField, TypeIntegerField, TypeInteger32Field, TypeBigIntegerField,
		TypePositiveBitField, TypePositiveSmallIntegerField, TypePositiveIntegerField, TypePositiveInteger32Field, TypePositiveBigIntegerField:
		switch val.Type().Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			if !ret.IsValid() {
				ret = val
			} else {
				ret.SetInt(val.Int())
			}
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			if !ret.IsValid() {
				ret = val
			} else {
				ret.SetUint(val.Uint())
			}
		default:
			msg := fmt.Sprintf("unexception value, exception type:[%d], current value:[%v]", typeValue, val.Interface())
			panic(msg)
		}
	case TypeFloatField, TypeDoubleField:
		switch val.Type().Kind() {
		case reflect.Float32, reflect.Float64:
			if !ret.IsValid() {
				ret = val
			} else {
				ret.SetFloat(val.Float())
			}
		default:
			msg := fmt.Sprintf("unexception value, exception type:[%d], current value:[%v]", typeValue, val.Interface())
			panic(msg)
		}
	case TypeStringField:
		switch val.Type().Kind() {
		case reflect.String:
			if !ret.IsValid() {
				ret = val
			} else {
				ret.SetString(val.String())
			}
		default:
			msg := fmt.Sprintf("unexception value, exception type:[%d], current value:[%v]", typeValue, val.Interface())
			panic(msg)
		}
	case TypeStructField:
		switch val.Type().Kind() {
		case reflect.Struct:
			if !ret.IsValid() {
				ret = val
			} else {
				ret.Set(val)
			}
		default:
			msg := fmt.Sprintf("unexception value, exception type:[%d], current value:[%v]", typeValue, val.Interface())
			panic(msg)
		}
	default:
		msg := fmt.Sprintf("unexception value, exception type:[%d], current value:[%v]", typeValue, val.Interface())
		panic(msg)
	}

	return
}
