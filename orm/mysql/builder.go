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
	str := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", s.getTableName(s.structInfo), s.getFieldNames(s.structInfo, false), s.getFieldValues(s.structInfo))
	log.Print(str)

	return str, nil
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
		ft := field.GetFieldTag()
		if ft.IsAutoIncrement() && !all {
			continue
		}

		if str == "" {
			str = fmt.Sprintf("`%s`", ft.Name())
		} else {
			str = fmt.Sprintf("%s,`%s`", str, ft.Name())
		}
	}

	return str
}

func (s *Builder) getFieldValues(info *model.StructInfo) string {
	str := ""
	for _, field := range *s.structInfo.GetFields() {
		ft := field.GetFieldTag()
		if ft.IsAutoIncrement() {
			continue
		}

		fValue := field.GetFieldValue()
		if str == "" {
			str = fmt.Sprintf("%s", fValue.GetValueStr())
		} else {
			str = fmt.Sprintf("%s,%s", str, fValue.GetValueStr())
		}
	}

	return str
}
