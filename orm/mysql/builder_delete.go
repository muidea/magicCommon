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
func (s *Builder) BuildDeleteRelation(fieldName string, relationInfo model.StructInfo) (ret string, err error) {
	leftVal, rightVal, errVal := s.getRelationValue(relationInfo)
	if errVal != nil {
		err = errVal
		return
	}

	ret = fmt.Sprintf("DELETE FROM `%s` WHERE `left`=%s AND `right`=%s", s.GetRelationTableName(fieldName, relationInfo), leftVal, rightVal)
	log.Print(ret)

	return
}
