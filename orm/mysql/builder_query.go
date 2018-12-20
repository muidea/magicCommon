package mysql

import (
	"fmt"
	"log"

	"muidea.com/magicCommon/orm/model"
)

// BuildQuery BuildQuery
func (s *Builder) BuildQuery() (ret string, err error) {
	pk := s.structInfo.GetPrimaryKey()
	if pk == nil {
		err = fmt.Errorf("no define primaryKey")
		return
	}

	pkfValue := pk.GetFieldValue()
	pkfTag := pk.GetFieldTag()
	pkfStr, pkferr := pkfValue.GetValueStr()
	if pkferr == nil {
		ret = fmt.Sprintf("SELECT %s FROM `%s` WHERE `%s`=%s", s.getFieldNames(s.structInfo, true), s.getTableName(s.structInfo), pkfTag.Name(), pkfStr)
		log.Print(ret)
	}
	err = pkferr

	return
}

// BuildQueryRelation BuildQueryRelation
func (s *Builder) BuildQueryRelation(relationInfo *model.StructInfo) (string, error) {
	str := "\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`left` INT NOT NULL,\n\t`right` INT NOT NULL,\n\tPRIMARY KEY (`id`)"
	str = fmt.Sprintf("CREATE TABLE `%s` (\n%s\n)\n", s.GetRelationTableName(relationInfo), str)
	log.Print(str)

	return str, nil
}
