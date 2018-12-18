package mysql

import (
	"fmt"

	"muidea.com/magicCommon/orm/model"
	"muidea.com/magicCommon/orm/util"
)

func verifyFieldInfo(fieldInfo *model.FieldInfo) error {
	tag := fieldInfo.GetFieldTag()
	ft := fieldInfo.GetFieldTag()
	if IsKeyWord(ft.Name()) {
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
	ft := fieldInfo.GetFieldTag()
	if ft.IsAutoIncrement() {
		autoIncrement = "AUTO_INCREMENT"
	}

	str := fmt.Sprintf("`%s` %s NOT NULL %s", ft.Name(), getFieldType(fieldInfo), autoIncrement)
	return str
}

func getFieldType(info *model.FieldInfo) (ret string) {
	ft := info.GetFieldType()
	switch ft.Value() {
	case util.TypeBooleanField:
		ret = "TINYINT"
		break
	case util.TypeStringField:
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
		msg := fmt.Sprintf("no support fileType, %d", ft.Value())
		panic(msg)
	}

	return
}
