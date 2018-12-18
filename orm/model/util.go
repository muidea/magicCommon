package model

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicCommon/orm/util"
)

// GetFieldType return field type as type constant from reflect.Value
func GetFieldType(val reflect.Type) (ft int, err error) {
	switch val.Kind() {
	case reflect.Int8:
		ft = util.TypeBitField
	case reflect.Uint8:
		ft = util.TypePositiveBitField
	case reflect.Int16:
		ft = util.TypeSmallIntegerField
	case reflect.Uint16:
		ft = util.TypePositiveSmallIntegerField
	case reflect.Int32:
		ft = util.TypeInteger32Field
	case reflect.Uint32:
		ft = util.TypePositiveInteger32Field
	case reflect.Int64:
		ft = util.TypeBigIntegerField
	case reflect.Uint64:
		ft = util.TypePositiveBigIntegerField
	case reflect.Int:
		ft = util.TypeIntegerField
	case reflect.Uint:
		ft = util.TypePositiveIntegerField
	case reflect.Float32:
		ft = util.TypeFloatField
	case reflect.Float64:
		ft = util.TypeDoubleField
	case reflect.Bool:
		ft = util.TypeBooleanField
	case reflect.String:
		ft = util.TypeStringField
	case reflect.Struct:
		switch val.String() {
		case "time.Time":
			ft = util.TypeDateTimeField
		default:
			ft = util.TypeStructField
		}
	case reflect.Slice:
		ft = util.TypeSliceField
	default:
		err = fmt.Errorf("unsupport field type:[%v], may be miss setting tag", val.Elem().Kind())
	}

	return
}

// IsReference IsReference
func IsReference(val reflect.Type) bool {
	switch val.Kind() {
	case reflect.Int8, reflect.Uint8, reflect.Int16, reflect.Uint16, reflect.Int32, reflect.Uint32,
		reflect.Int64, reflect.Uint64, reflect.Int, reflect.Uint, reflect.Float32, reflect.Float64, reflect.Bool, reflect.String:
		return false
	case reflect.Struct:
		switch val.String() {
		case "time.Time":
			return false
		default:
			return true
		}
	case reflect.Slice:
		val = val.Elem()
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		switch val.Kind() {
		case reflect.Int8, reflect.Uint8, reflect.Int16, reflect.Uint16, reflect.Int32, reflect.Uint32,
			reflect.Int64, reflect.Uint64, reflect.Int, reflect.Uint, reflect.Float32, reflect.Float64, reflect.Bool, reflect.String:
			return false
		case reflect.Struct:
			switch val.String() {
			case "time.Time":
				return false
			default:
				return true
			}
		default:
			err := fmt.Errorf("unsupport slice field type:[%v], may be miss setting tag", val.Kind())
			panic(err.Error())
		}
	default:
		err := fmt.Errorf("unsupport field type:[%v], may be miss setting tag", val.Kind())
		panic(err.Error())
	}
}

// GetStructValue GetStructValue
func GetStructValue(val reflect.Value) (ret FieldValue) {
	log.Print("$$$$$$$$$")
	log.Print(val)
	structInfo, _ := getStructInfo(reflect.Indirect(val))
	pk := structInfo.GetPrimaryKey()
	if pk == nil {
		panic("illegal structVal, no define PrimaryKey")
	}

	ret = pk.GetFieldValue()
	return
}
