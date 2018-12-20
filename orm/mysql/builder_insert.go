package mysql

import (
	"fmt"
	"log"

	"muidea.com/magicCommon/orm/model"
)

// BuildInsert  BuildInsert
func (s *Builder) BuildInsert() (ret string, err error) {
	sql := ""
	vals, verr := s.getFieldValues(s.structInfo)
	if verr == nil {
		for _, val := range vals {
			sql = fmt.Sprintf("%sINSERT INTO `%s` (%s) VALUES (%s);", sql, s.getTableName(s.structInfo), s.getFieldNames(s.structInfo, false), val)
		}
		log.Print(sql)
		ret = sql
	}
	err = verr

	return
}

// BuildInsertRelation BuildInsertRelation
func (s *Builder) BuildInsertRelation(relationInfo *model.StructInfo) (ret string, err error) {
	leftVal, rightVal, errVal := s.getRelationValue(relationInfo)
	if errVal != nil {
		err = errVal
		return
	}

	ret = fmt.Sprintf("INSERT INTO `%s` (`left`, `right`) VALUES (%s,%s);", s.GetRelationTableName(relationInfo), leftVal, rightVal)
	log.Print(ret)

	return
}
