package util

import (
	"fmt"
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

// Field type
const (
	TypeBaseTypeField = iota
	TypeReferenceField
	TypeReferencePtrField
)

// IsBaseTypeValue IsBaseTypeValue
func IsBaseTypeValue(typeValue int) bool {
	return typeValue < TypeStructField
}

// GetBaseTypeInitValue GetBaseTypeInitValue
func GetBaseTypeInitValue(typeValue int) (ret interface{}) {
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
