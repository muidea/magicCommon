package model

import (
	"fmt"
	"reflect"

	"muidea.com/magicCommon/orm"
)

// GetFieldType return field type as type constant from reflect.Value
func GetFieldType(val reflect.Type) (ft int, err error) {
	switch val.Kind() {
	case reflect.Int8:
		ft = orm.TypeBitField
	case reflect.Int16:
		ft = orm.TypeSmallIntegerField
	case reflect.Int32, reflect.Int:
		ft = orm.TypeIntegerField
	case reflect.Int64:
		ft = orm.TypeBigIntegerField
	case reflect.Uint8:
		ft = orm.TypePositiveBitField
	case reflect.Uint16:
		ft = orm.TypePositiveSmallIntegerField
	case reflect.Uint32, reflect.Uint:
		ft = orm.TypePositiveIntegerField
	case reflect.Uint64:
		ft = orm.TypePositiveBigIntegerField
	case reflect.Float32:
		ft = orm.TypeFloatField
	case reflect.Float64:
		ft = orm.TypeDoubleField
	case reflect.Bool:
		ft = orm.TypeBooleanField
	case reflect.String:
		ft = orm.TypeVarCharField
	case reflect.Struct:
		switch val.String() {
		case "time.Time":
			ft = orm.TypeDateTimeField
		default:
			ft = orm.TypeStrictField
		}
	default:
		err = fmt.Errorf("unsupport field type %v, may be miss setting tag", val.Name())
	}

	return
}
