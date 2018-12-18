package mysql

import (
	"fmt"
	"log"
	"strings"

	"muidea.com/magicCommon/orm/model"
)

// Builder Builder
type Builder struct {
	structInfo *model.StructInfo
}

// New create builder
func New(structInfo *model.StructInfo) *Builder {
	err := verifyStructInfo(structInfo)
	if err != nil {
		log.Printf("verify structInfo failed, err:%s", err.Error())
		return nil
	}

	return &Builder{structInfo: structInfo}
}

// BuildCreateSchema  BuildCreateSchema
func (s *Builder) BuildCreateSchema() (string, error) {
	str := ""
	for _, val := range *s.structInfo.GetFields() {
		if str == "" {
			str = fmt.Sprintf("\t%s", declareFieldInfo(val))
		} else {
			str = fmt.Sprintf("%s,\n\t%s", str, declareFieldInfo(val))
		}
	}
	if s.structInfo.GetPrimaryKey() != nil {
		fTag := s.structInfo.GetPrimaryKey().GetFieldTag()
		str = fmt.Sprintf("%s,\n\tPRIMARY KEY (`%s`)", str, fTag.Name())
	}

	str = fmt.Sprintf("CREATE TABLE `%s` (\n%s\n)\n", s.getTableName(s.structInfo), str)
	log.Print(str)

	return str, nil
}

// BuildDropSchema  BuildDropSchema
func (s *Builder) BuildDropSchema() (string, error) {
	str := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", s.getTableName(s.structInfo))
	log.Print(str)

	return str, nil
}

// BuildInsert  BuildInsert
func (s *Builder) BuildInsert() (string, error) {
	sql := ""
	vals := s.getFieldValues(s.structInfo)
	for _, val := range vals {
		sql = fmt.Sprintf("%sINSERT INTO `%s` (%s) VALUES (%s);", sql, s.getTableName(s.structInfo), s.getFieldNames(s.structInfo, false), val)
	}
	log.Print(sql)

	return sql, nil
}

// BuildUpdate  BuildUpdate
func (s *Builder) BuildUpdate() (string, error) {
	str := ""
	for _, val := range *s.structInfo.GetFields() {
		fValue := val.GetFieldValue()
		fTag := val.GetFieldTag()
		if val != s.structInfo.GetPrimaryKey() {
			if str == "" {
				str = fmt.Sprintf("`%s`=%s", fTag.Name(), fValue.GetValueStr())
			} else {
				str = fmt.Sprintf("%s,`%s`=%s", str, fTag.Name(), fValue.GetValueStr())
			}
		}
	}

	pkfValue := s.structInfo.GetPrimaryKey().GetFieldValue()
	pkfTag := s.structInfo.GetPrimaryKey().GetFieldTag()
	str = fmt.Sprintf("UPDATE `%s` SET %s WHERE `%s`=%s", s.getTableName(s.structInfo), str, pkfTag.Name(), pkfValue.GetValueStr())
	log.Print(str)

	return str, nil
}

// BuildDelete  BuildDelete
func (s *Builder) BuildDelete() (string, error) {
	pkfValue := s.structInfo.GetPrimaryKey().GetFieldValue()
	pkfTag := s.structInfo.GetPrimaryKey().GetFieldTag()
	str := fmt.Sprintf("DELETE FROM `%s` WHERE `%s`=%s", s.getTableName(s.structInfo), pkfTag.Name(), pkfValue.GetValueStr())
	log.Print(str)

	return str, nil
}

// BuildQuery BuildQuery
func (s *Builder) BuildQuery() (string, error) {
	pkfValue := s.structInfo.GetPrimaryKey().GetFieldValue()
	pkfTag := s.structInfo.GetPrimaryKey().GetFieldTag()
	str := fmt.Sprintf("SELECT %s FROM `%s` WHERE `%s`=%s", s.getFieldNames(s.structInfo, true), s.getTableName(s.structInfo), pkfTag.Name(), pkfValue.GetValueStr())
	log.Print(str)

	return str, nil
}

func (s *Builder) getTableName(info *model.StructInfo) string {
	return strings.Join(strings.Split(s.structInfo.GetStructName(), "."), "_")
}

func (s *Builder) getFieldNames(info *model.StructInfo, all bool) string {
	str := ""
	for _, field := range *s.structInfo.GetFields() {
		fTag := field.GetFieldTag()
		if fTag.IsAutoIncrement() && !all {
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

func (s *Builder) getFieldValues(info *model.StructInfo) (ret []string) {
	str := ""
	for _, field := range *info.GetFields() {
		fTag := field.GetFieldTag()
		if fTag.IsAutoIncrement() {
			continue
		}

		fType := field.GetFieldType()
		fValue := field.GetFieldValue()
		if fType.IsReference() {
			fValue = model.GetStructValue(fValue.GetValue())
		}

		if str == "" {
			str = fmt.Sprintf("%s", fValue.GetValueStr())
		} else {
			str = fmt.Sprintf("%s,%s", str, fValue.GetValueStr())
		}
	}

	ret = append(ret, str)

	return
}
