package mysql

import (
	"fmt"

	"muidea.com/magicCommon/orm/model"
	"muidea.com/magicCommon/orm/util"
)

func verifyFieldInfo(fieldInfo model.FieldInfo) error {
	fTag := fieldInfo.GetFieldTag()
	if IsKeyWord(fTag.Name()) {
		return fmt.Errorf("illegal fieldTag, is a key word.[%s]", fTag)
	}

	return nil
}

func verifyStructInfo(structInfo model.StructInfo) error {
	name := structInfo.GetName()
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

func declareFieldInfo(fieldInfo model.FieldInfo) string {
	autoIncrement := ""
	fTag := fieldInfo.GetFieldTag()
	if fTag.IsAutoIncrement() {
		autoIncrement = "AUTO_INCREMENT"
	}

	allowNull := "NOT NULL"
	fType := fieldInfo.GetFieldType()
	if fType.IsPtr() {
		allowNull = ""
	}

	str := fmt.Sprintf("`%s` %s %s %s", fTag.Name(), getFieldType(fieldInfo), allowNull, autoIncrement)
	return str
}

func getFieldType(info model.FieldInfo) (ret string) {
	fType := info.GetFieldType()
	switch fType.Value() {
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
	case util.TypeSliceField:
		ret = "TEXT"
	default:
		msg := fmt.Sprintf("no support fileType, %d", fType.Value())
		panic(msg)
	}

	return
}
