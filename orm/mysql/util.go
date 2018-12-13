package mysql

import (
	"fmt"

	"muidea.com/magicCommon/orm"
	"muidea.com/magicCommon/orm/model"
)

func verifyFieldInfo(fieldInfo *model.FieldInfo) error {
	tag := fieldInfo.GetFieldTag()
	if IsKeyWord(tag) {
		return fmt.Errorf("illegal fieldTag, is a key word.[%s]", tag)
	}

	return nil
}

func verifyStructInfo(structInfo *model.StructInfo) error {
	name := structInfo.GetStructName()
	if IsKeyWord(name) {
		return fmt.Errorf("illegal structName, is a key word.[%s]", name)
	}

	for _, val := range structInfo.Fields.Fields {
		err := verifyFieldInfo(val)
		if err != nil {
			return err
		}
	}

	return nil
}

func declareFieldInfo(fieldInfo *model.FieldInfo) string {
	autoIncrement := ""
	if fieldInfo.IsAutoIncrement() {
		autoIncrement = "AUTO_INCREMENT"
	}

	if fieldInfo.GetFieldTypeValue() < orm.TypeStrictField {
		str := fmt.Sprintf("`%s` %s NOT NULL %s", fieldInfo.GetFieldTag(), getFieldType(fieldInfo), autoIncrement)
		return str
	}

	return ""
}

func getFieldType(info *model.FieldInfo) (ret string) {
	typeValue := info.GetFieldTypeValue()
	switch typeValue {
	case orm.TypeBooleanField:
		ret = "TINYINT"
		break
	case orm.TypeVarCharField:
		ret = "TEXT"
		break
	case orm.TypeDateTimeField:
		ret = "DATETIME"
		break
	case orm.TypeBitField:
		ret = "TINYINT"
		break
	case orm.TypeSmallIntegerField:
		ret = "SMALLINT"
		break
	case orm.TypeIntegerField:
		ret = "INT"
		break
	case orm.TypeBigIntegerField:
		ret = "BIGINT"
		break
	case orm.TypePositiveBitField:
		ret = "SMALLINT"
		break
	case orm.TypePositiveSmallIntegerField:
		ret = "INT"
		break
	case orm.TypePositiveIntegerField:
		ret = "BIGINT"
		break
	case orm.TypePositiveBigIntegerField:
		ret = "BIGINT"
		break
	case orm.TypeFloatField:
		ret = "FLOAT"
		break
	case orm.TypeDoubleField:
		ret = "DOUBLE"
		break
	default:
		msg := fmt.Sprintf("no support fileType, %d", typeValue)
		panic(msg)
	}

	return
}
