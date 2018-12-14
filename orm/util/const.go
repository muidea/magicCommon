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
)

// GetEmptyValue GetEmptyValue
func GetEmptyValue(typeValue int) (ret interface{}) {
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
