package orm

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
	TypeDecimalField
	TypeJSONField
	TypeJsonbField
	RelForeignKey
	RelOneToOne
	RelManyToMany
	RelReverseOne
	RelReverseMany
)

// Define some logic enum
const (
	IsIntegerField         = ^-TypePositiveBigIntegerField >> 6 << 7
	IsPositiveIntegerField = ^-TypePositiveBigIntegerField >> 10 << 11
	IsRelField             = ^-RelReverseMany >> 18 << 19
	IsFieldType            = ^-RelReverseMany<<1 + 1
)
