package mysql

import (
	"fmt"
	"strings"

	"muidea.com/magicCommon/orm/model"
)

// Builder Builder
type Builder struct {
	structInfo *model.StructInfo
}

// New create builder
func New(structInfo *model.StructInfo) *Builder {
	//err := verifyStructInfo(structInfo)
	//if err != nil {
	//	log.Printf("verify structInfo failed, err:%s", err.Error())
	//	return nil
	//}

	return &Builder{structInfo: structInfo}
}

func (s *Builder) getTableName(info *model.StructInfo) string {
	return strings.Join(strings.Split(info.GetStructName(), "."), "_")
}

// GetTableName GetTableName
func (s *Builder) GetTableName() string {
	return s.getTableName(s.structInfo)
}

func (s *Builder) getFieldNames(info *model.StructInfo, all bool) string {
	str := ""
	for _, field := range *s.structInfo.GetFields() {
		fTag := field.GetFieldTag()
		if fTag.IsAutoIncrement() && !all {
			continue
		}

		fType := field.GetFieldType()
		fValue := field.GetFieldValue()
		if fType.IsPtr() && fValue.IsNil() {
			continue
		}

		if str == "" {
			str = fmt.Sprintf("`%s`", fTag.Name())
		} else {
			str = fmt.Sprintf("%s,`%s`", str, fTag.Name())
		}
	}

	return str
}

// GetRelationTableName GetRelationTableName
func (s *Builder) GetRelationTableName(relationInfo *model.StructInfo) string {
	leftName := s.getTableName(s.structInfo)
	rightName := s.getTableName(relationInfo)

	if strings.Compare(leftName, rightName) < 0 {
		return fmt.Sprintf("%s2%s", leftName, rightName)
	}

	return fmt.Sprintf("%s2%s", rightName, leftName)
}

func (s *Builder) getRelationValue(relationInfo *model.StructInfo) (leftVal, rightVal string, err error) {
	leftName := s.getTableName(s.structInfo)
	rightName := s.getTableName(relationInfo)

	structKey := s.structInfo.GetPrimaryKey()
	relationKey := relationInfo.GetPrimaryKey()
	if structKey == nil || relationKey == nil {
		err = fmt.Errorf("no define primaryKey")
		return
	}

	structVal, structErr := structKey.GetFieldValue().GetValueStr()
	if structErr != nil {
		err = structErr
		return
	}
	relationVal, relationErr := relationKey.GetFieldValue().GetValueStr()
	if relationErr != nil {
		err = relationErr
		return
	}

	if strings.Compare(leftName, rightName) < 0 {
		leftVal = structVal
		rightVal = relationVal
		return
	}

	leftVal = relationVal
	rightVal = structVal
	return
}
