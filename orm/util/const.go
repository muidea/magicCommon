package util

import (
	"fmt"
	"time"
)

// Define the Type enum
const (
	TypeBooleanField = 1 << iota
	TypeVarCharField
	TypeCharField
	TypeTextField
	TypeTimeField
	TypeDateField
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
	TypeDecimalField
	TypeStrictField
)

// GetEmptyValue GetEmptyValue
func GetEmptyValue(typeValue int) (ret interface{}) {
	switch typeValue {
	case TypeBooleanField:
		val := 0
		ret = &val
		break
	case TypeVarCharField:
		val := ""
		ret = &val
		break
	case TypeDateTimeField:
		val := time.Time{}
		ret = &val
		break
	case TypeBitField:
		val := int8(0)
		ret = &val
		break
	case TypeSmallIntegerField:
		val := int16(0)
		ret = &val
		break
	case TypeIntegerField:
		val := int(0)
		ret = &val
		break
	case TypeInteger32Field:
		val := int32(0)
		ret = &val
		break
	case TypeBigIntegerField:
		val := int64(0)
		ret = &val
		break
	case TypePositiveBitField:
		val := uint8(0)
		ret = &val
		break
	case TypePositiveSmallIntegerField:
		val := uint16(0)
		ret = &val
		break
	case TypePositiveIntegerField:
		val := uint(0)
		ret = &val
		break
	case TypePositiveInteger32Field:
		val := uint32(0)
		ret = &val
		break
	case TypePositiveBigIntegerField:
		val := uint64(0)
		ret = &val
		break
	case TypeFloatField:
		val := float32(0.00)
		ret = &val
		break
	case TypeDoubleField:
		val := float64(0.00)
		ret = &val
		break
	default:
		msg := fmt.Sprintf("no support fileType, %d", typeValue)
		panic(msg)
	}
	return
}
