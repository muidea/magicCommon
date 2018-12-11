package orm

import (
	"fmt"
	"reflect"
)

// return field type as type constant from reflect.Value
func getFieldType(val reflect.Type) (ft int, err error) {
	switch val.Kind() {
	case reflect.Int8:
		ft = TypeBitField
	case reflect.Int16:
		ft = TypeSmallIntegerField
	case reflect.Int32, reflect.Int:
		ft = TypeIntegerField
	case reflect.Int64:
		ft = TypeBigIntegerField
	case reflect.Uint8:
		ft = TypePositiveBitField
	case reflect.Uint16:
		ft = TypePositiveSmallIntegerField
	case reflect.Uint32, reflect.Uint:
		ft = TypePositiveIntegerField
	case reflect.Uint64:
		ft = TypePositiveBigIntegerField
	case reflect.Float32, reflect.Float64:
		ft = TypeFloatField
	case reflect.Bool:
		ft = TypeBooleanField
	case reflect.String:
		ft = TypeVarCharField
	case reflect.Struct:
		ft = TypeStrictField
	case reflect.Ptr:
		ft = TypePtrField
	default:
		err = fmt.Errorf("unsupport field type %v, may be miss setting tag", val.Name())
	}

	return
}
