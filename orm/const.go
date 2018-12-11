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
	TypeStrictField
	TypeJSONField
	TypeJsonbField
	RelForeignKey
	RelOneToOne
	RelManyToMany
	RelReverseOne
	RelReverseMany
)
