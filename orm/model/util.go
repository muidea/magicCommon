package model

import (
	"fmt"
	"reflect"

	"muidea.com/magicCommon/orm/util"
)

// GetFieldType return field type as type constant from reflect.Value
func GetFieldType(val reflect.Type) (ft int, err error) {
	switch val.Kind() {
	case reflect.Int8:
		ft = util.TypeBitField
	case reflect.Int16:
		ft = util.TypeSmallIntegerField
	case reflect.Int32, reflect.Int:
		ft = util.TypeIntegerField
	case reflect.Int64:
		ft = util.TypeBigIntegerField
	case reflect.Uint8:
		ft = util.TypePositiveBitField
	case reflect.Uint16:
		ft = util.TypePositiveSmallIntegerField
	case reflect.Uint32, reflect.Uint:
		ft = util.TypePositiveIntegerField
	case reflect.Uint64:
		ft = util.TypePositiveBigIntegerField
	case reflect.Float32:
		ft = util.TypeFloatField
	case reflect.Float64:
		ft = util.TypeDoubleField
	case reflect.Bool:
		ft = util.TypeBooleanField
	case reflect.String:
		ft = util.TypeVarCharField
	case reflect.Struct:
		switch val.String() {
		case "time.Time":
			ft = util.TypeDateTimeField
		default:
			ft = util.TypeStrictField
		}
	default:
		err = fmt.Errorf("unsupport field type %v, may be miss setting tag", val.Name())
	}

	return
}
