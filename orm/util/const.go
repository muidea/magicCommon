package util

import (
	"fmt"
	"reflect"
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
)

// IsBasicType IsBasicType
func IsBasicType(typeValue int) bool {
	return typeValue < TypeStructField
}

// IsStructType IsStructType
func IsStructType(typeValue int) bool {
	return typeValue == TypeStructField
}

// IsSliceType IsSliceType
func IsSliceType(typeValue int) bool {
	return typeValue == TypeSliceField
}

// GetBasicTypeInitValue GetBasicTypeInitValue
func GetBasicTypeInitValue(typeValue int) (ret interface{}) {
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
	case TypeStructField:
		val := 0
		ret = &val
	case TypeSliceField:
		val := ""
		ret = &val
	default:
		msg := fmt.Sprintf("no support fileType, %d", typeValue)
		panic(msg)
	}
	return
}

// GetTypeValueEnum return field type as type constant from reflect.Value
func GetTypeValueEnum(val reflect.Type) (ft int, err error) {
	switch val.Kind() {
	case reflect.Int8:
		ft = TypeBitField
	case reflect.Uint8:
		ft = TypePositiveBitField
	case reflect.Int16:
		ft = TypeSmallIntegerField
	case reflect.Uint16:
		ft = TypePositiveSmallIntegerField
	case reflect.Int32:
		ft = TypeInteger32Field
	case reflect.Uint32:
		ft = TypePositiveInteger32Field
	case reflect.Int64:
		ft = TypeBigIntegerField
	case reflect.Uint64:
		ft = TypePositiveBigIntegerField
	case reflect.Int:
		ft = TypeIntegerField
	case reflect.Uint:
		ft = TypePositiveIntegerField
	case reflect.Float32:
		ft = TypeFloatField
	case reflect.Float64:
		ft = TypeDoubleField
	case reflect.Bool:
		ft = TypeBooleanField
	case reflect.String:
		ft = TypeStringField
	case reflect.Struct:
		switch val.String() {
		case "time.Time":
			ft = TypeDateTimeField
		default:
			ft = TypeStructField
		}
	case reflect.Slice:
		ft = TypeSliceField
	default:
		err = fmt.Errorf("unsupport field type:[%v], may be miss setting tag", val.String())
	}

	return
}

// GetSliceRawTypeEnum get slice rawType
func GetSliceRawTypeEnum(sliceType reflect.Type) (ret int, err error) {
	if sliceType.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal type, not slice. typeVal:%s", sliceType.Kind().String())
		return
	}

	rawType := sliceType.Elem()
	if rawType.Kind() == reflect.Ptr {
		rawType = rawType.Elem()
	}
	ret, err = GetTypeValueEnum(rawType)
	if err != nil {
		return
	}

	return
}
