package mysql

import (
	"fmt"

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
