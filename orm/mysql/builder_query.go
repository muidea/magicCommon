package mysql

import (
	"fmt"
	"log"

	"muidea.com/magicCommon/orm/model"
)

// BuildQuery BuildQuery
func (s *Builder) BuildQuery() (ret string, err error) {
	pk := s.structInfo.GetPrimaryField()
	if pk == nil {
		err = fmt.Errorf("no define primaryKey")
		return
	}

	pkfValue := pk.GetFieldValue()
	pkfTag := pk.GetFieldTag()
	pkfStr, pkferr := pkfValue.GetValueStr()
	if pkferr == nil {
		ret = fmt.Sprintf("SELECT %s FROM `%s` WHERE `%s`=%s", s.getFieldQueryNames(s.structInfo), s.getTableName(s.structInfo), pkfTag.Name(), pkfStr)
		log.Print(ret)
	}
	err = pkferr

	return
}

// BuildQueryRelation BuildQueryRelation
func (s *Builder) BuildQueryRelation(fieldName string, relationInfo model.StructInfo) (ret string, err error) {
	pk := s.structInfo.GetPrimaryField()
	if pk == nil {
		err = fmt.Errorf("no define primaryKey")
		return
	}

	pkfValue := pk.GetFieldValue()
	pkfStr, pkferr := pkfValue.GetValueStr()
	if pkferr == nil {
		ret = fmt.Sprintf("SELECT `right` FROM `%s` WHERE `left`= %s", s.GetRelationTableName(fieldName, relationInfo), pkfStr)
		log.Print(ret)
	}

	err = pkferr

	return
}

func (s *Builder) getFieldQueryNames(info model.StructInfo) string {
	str := ""
	for _, field := range *s.structInfo.GetFields() {
		fTag := field.GetFieldTag()
		fType := field.GetFieldType()

		dependType, _ := fType.Depend()
		if dependType != nil {
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
