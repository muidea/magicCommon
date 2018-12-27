package mysql

import (
	"fmt"
	"log"

	"muidea.com/magicCommon/orm/model"
	"muidea.com/magicCommon/orm/util"
)

// BuildUpdate  BuildUpdate
func (s *Builder) BuildUpdate() (ret string, err error) {
	str := ""
	for _, val := range *s.structInfo.GetFields() {
		fType := val.GetFieldType()
		fValue := val.GetFieldValue()
		fTag := val.GetFieldTag()
		if fType.IsPtr() && fValue.IsNil() {
			continue
		}

		if !util.IsBasicType(fType.Value()) {
			continue
		}

		if val != s.structInfo.GetPrimaryKey() {
			fStr, ferr := fValue.GetValueStr()
			if ferr != nil {
				err = ferr
				break
			}
			if str == "" {
				str = fmt.Sprintf("`%s`=%s", fTag.Name(), fStr)
			} else {
				str = fmt.Sprintf("%s,`%s`=%s", str, fTag.Name(), fStr)
			}
		}
	}

	if err != nil {
		return
	}

	pkfValue := s.structInfo.GetPrimaryKey().GetFieldValue()
	pkfTag := s.structInfo.GetPrimaryKey().GetFieldTag()
	pkfStr, pkferr := pkfValue.GetValueStr()
	if pkferr == nil {
		str = fmt.Sprintf("UPDATE `%s` SET %s WHERE `%s`=%s", s.getTableName(s.structInfo), str, pkfTag.Name(), pkfStr)
		log.Print(str)
	}

	ret = str
	err = pkferr

	return
}

// BuildUpdateRelation BuildUpdateRelation
func (s *Builder) BuildUpdateRelation(relationInfo *model.StructInfo) (string, error) {
	str := "\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`left` INT NOT NULL,\n\t`right` INT NOT NULL,\n\tPRIMARY KEY (`id`)"
	str = fmt.Sprintf("CREATE TABLE `%s` (\n%s\n)\n", s.GetRelationTableName(relationInfo), str)
	log.Print(str)

	return str, nil
}
