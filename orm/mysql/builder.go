package mysql

import (
	"fmt"
	"strings"

	"muidea.com/magicCommon/orm/model"
)

// Builder Builder
type Builder struct {
	structInfo model.StructInfo
}

// New create builder
func New(structInfo model.StructInfo) *Builder {
	//err := verifyStructInfo(structInfo)
	//if err != nil {
	//	log.Printf("verify structInfo failed, err:%s", err.Error())
	//	return nil
	//}

	return &Builder{structInfo: structInfo}
}

func (s *Builder) getTableName(info model.StructInfo) string {
	return strings.Join(strings.Split(info.GetName(), "."), "_")
}

// GetTableName GetTableName
func (s *Builder) GetTableName() string {
	return s.getTableName(s.structInfo)
}

// GetRelationTableName GetRelationTableName
func (s *Builder) GetRelationTableName(fieldName string, relationInfo model.StructInfo) string {
	leftName := s.getTableName(s.structInfo)
	rightName := s.getTableName(relationInfo)

	return fmt.Sprintf("%s%s2%s", leftName, fieldName, rightName)
}

func (s *Builder) getRelationValue(relationInfo model.StructInfo) (leftVal, rightVal string, err error) {
	structKey := s.structInfo.GetPrimaryField()
	relationKey := relationInfo.GetPrimaryField()
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

	leftVal = structVal
	rightVal = relationVal
	return
}
