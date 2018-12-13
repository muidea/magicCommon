package mysql

import (
	"fmt"

	"muidea.com/magicCommon/orm/model"
	"muidea.com/magicCommon/orm/util"
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

	for _, val := range *structInfo.GetFields() {
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

	if fieldInfo.GetFieldTypeValue() < util.TypeStrictField {
		str := fmt.Sprintf("`%s` %s NOT NULL %s", fieldInfo.GetFieldTag(), getFieldType(fieldInfo), autoIncrement)
		return str
	}

	return ""
}

func getFieldType(info *model.FieldInfo) (ret string) {
	typeValue := info.GetFieldTypeValue()
	switch typeValue {
	case util.TypeBooleanField:
		ret = "TINYINT"
		break
	case util.TypeVarCharField:
		ret = "TEXT"
		break
	case util.TypeDateTimeField:
		ret = "DATETIME"
		break
	case util.TypeBitField:
		ret = "TINYINT"
		break
	case util.TypeSmallIntegerField:
		ret = "SMALLINT"
		break
	case util.TypeIntegerField:
		ret = "INT"
		break
	case util.TypeInteger32Field:
		ret = "INT"
		break
	case util.TypeBigIntegerField:
		ret = "BIGINT"
		break
	case util.TypePositiveBitField:
		ret = "SMALLINT"
		break
	case util.TypePositiveSmallIntegerField:
		ret = "INT"
		break
	case util.TypePositiveIntegerField:
		ret = "BIGINT"
		break
	case util.TypePositiveInteger32Field:
		ret = "BIGINT"
		break
	case util.TypePositiveBigIntegerField:
		ret = "BIGINT"
		break
	case util.TypeFloatField:
		ret = "FLOAT"
		break
	case util.TypeDoubleField:
		ret = "DOUBLE"
		break
	default:
		msg := fmt.Sprintf("no support fileType, %d", typeValue)
		panic(msg)
	}

	return
}
