package util

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
	TypeIntegerField
	TypeBigIntegerField
	TypePositiveBitField
	TypePositiveSmallIntegerField
	TypePositiveIntegerField
	TypePositiveBigIntegerField
	TypeFloatField
	TypeDoubleField
	TypeDecimalField
	TypeStrictField
)
