package mysql

import (
	"fmt"
	"log"

	"muidea.com/magicCommon/orm/model"
	"muidea.com/magicCommon/orm/util"
)

// BuildInsert  BuildInsert
func (s *Builder) BuildInsert() (ret string, err error) {
	sql := ""
	vals, verr := s.getFieldInsertValues(s.structInfo)
	if verr == nil {
		for _, val := range vals {
			sql = fmt.Sprintf("%sINSERT INTO `%s` (%s) VALUES (%s);", sql, s.getTableName(s.structInfo), s.getFieldInsertNames(s.structInfo), val)
		}
		log.Print(sql)
		ret = sql
	}
	err = verr

	return
}

// BuildInsertRelation BuildInsertRelation
func (s *Builder) BuildInsertRelation(fieldName string, relationInfo model.StructInfo) (ret string, err error) {
	leftVal, rightVal, errVal := s.getRelationValue(relationInfo)
	if errVal != nil {
		err = errVal
		return
	}

	ret = fmt.Sprintf("INSERT INTO `%s` (`left`, `right`) VALUES (%s,%s);", s.GetRelationTableName(fieldName, relationInfo), leftVal, rightVal)
	log.Print(ret)

	return
}

func (s *Builder) getFieldInsertNames(info model.StructInfo) string {
	str := ""
	for _, field := range *s.structInfo.GetFields() {
		fTag := field.GetFieldTag()
		if fTag.IsAutoIncrement() {
			continue
		}

		fType := field.GetFieldType()
		fValue := field.GetFieldValue()
		if fType.IsPtr() && fValue.IsNil() {
			continue
		}

		if !util.IsBasicType(fType.Value()) {
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

func (s *Builder) getFieldInsertValues(info model.StructInfo) (ret []string, err error) {
	str := ""
	for _, field := range *info.GetFields() {
		fTag := field.GetFieldTag()
		if fTag.IsAutoIncrement() {
			continue
		}

		fType := field.GetFieldType()
		fValue := field.GetFieldValue()
		if fType.IsPtr() && fValue.IsNil() {
			continue
		}

		if !util.IsBasicType(fType.Value()) {
			continue
		}

		fStr, ferr := fValue.GetValueStr()
		if ferr == nil {
			if str == "" {
				str = fmt.Sprintf("%s", fStr)
			} else {
				str = fmt.Sprintf("%s,%s", str, fStr)
			}
		} else {
			err = ferr
			break
		}
	}

	ret = append(ret, str)

	return
}
