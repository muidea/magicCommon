package mysql

import (
	"fmt"
	"log"

	"muidea.com/magicCommon/orm/model"
)

// BuildDelete  BuildDelete
func (s *Builder) BuildDelete() (ret string, err error) {
	pkfValue := s.structInfo.GetPrimaryField().GetFieldValue()
	pkfTag := s.structInfo.GetPrimaryField().GetFieldTag()
	pkfStr, pkferr := pkfValue.GetValueStr()
	if pkferr == nil {
		ret = fmt.Sprintf("DELETE FROM `%s` WHERE `%s`=%s", s.getTableName(s.structInfo), pkfTag.Name(), pkfStr)
		log.Print(ret)
	}

	err = pkferr

	return
}

// BuildDeleteRelation BuildDeleteRelation
func (s *Builder) BuildDeleteRelation(fieldName string, relationInfo model.StructInfo) (delRight, delRelation string, err error) {
	leftVal, leftErr := s.getStructValue(s.structInfo)
	if leftErr != nil {
		err = leftErr
		return
	}

	delRight = fmt.Sprintf("DELETE FROM `%s` WHERE `id` in (SELECT `right` FROM `%s` WHERE `left`=%s)", s.getTableName(relationInfo), s.GetRelationTableName(fieldName, relationInfo), leftVal)
	log.Print(delRight)

	delRelation = fmt.Sprintf("DELETE FROM `%s` WHERE `left`=%s", s.GetRelationTableName(fieldName, relationInfo), leftVal)
	log.Print(delRelation)

	return
}
