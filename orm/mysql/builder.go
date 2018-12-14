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
		str = fmt.Sprintf("%s,\n\tPRIMARY KEY (`%s`)", str, (s.structInfo.GetPrimaryKey().GetFieldTag()))
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
		if val != s.structInfo.GetPrimaryKey() {
			if str == "" {
				str = fmt.Sprintf("`%s`=%s", val.GetFieldTag(), val.GetFieldValueStr())
			} else {
				str = fmt.Sprintf("%s,`%s`=%s", str, val.GetFieldTag(), val.GetFieldValueStr())
			}
		}
	}

	str = fmt.Sprintf("UPDATE `%s` SET %s WHERE `%s`=%s", s.getTableName(s.structInfo), str, s.structInfo.GetPrimaryKey().GetFieldTag(), s.structInfo.GetPrimaryKey().GetFieldValueStr())
	log.Print(str)

	return str, nil
}

// BuildDelete  BuildDelete
func (s *Builder) BuildDelete() (string, error) {
	str := fmt.Sprintf("DELETE FROM `%s` WHERE `%s`=%s", s.getTableName(s.structInfo), s.structInfo.GetPrimaryKey().GetFieldTag(), s.structInfo.GetPrimaryKey().GetFieldValueStr())
	log.Print(str)

	return str, nil
}

// BuildQuery BuildQuery
func (s *Builder) BuildQuery() (string, error) {
	str := fmt.Sprintf("SELECT %s FROM `%s` WHERE `%s`=%s", s.getFieldNames(s.structInfo, true), s.getTableName(s.structInfo), s.structInfo.GetPrimaryKey().GetFieldTag(), s.structInfo.GetPrimaryKey().GetFieldValueStr())
	log.Print(str)

	return str, nil
}

func (s *Builder) getTableName(info *model.StructInfo) string {
	return strings.Join(strings.Split(s.structInfo.GetStructName(), "."), "_")
}

func (s *Builder) getFieldNames(info *model.StructInfo, all bool) string {
	str := ""
	for _, field := range *s.structInfo.GetFields() {
		if field.IsAutoIncrement() && !all {
			continue
		}

		if str == "" {
			str = fmt.Sprintf("`%s`", field.GetFieldTag())
		} else {
			str = fmt.Sprintf("%s,`%s`", str, field.GetFieldTag())
		}
	}

	return str
}

func (s *Builder) getFieldValues(info *model.StructInfo) string {
	str := ""
	for _, field := range *s.structInfo.GetFields() {
		if field.IsAutoIncrement() {
			continue
		}

		if str == "" {
			str = fmt.Sprintf("%s", field.GetFieldValueStr())
		} else {
			str = fmt.Sprintf("%s,%s", str, field.GetFieldValueStr())
		}
	}

	return str
}
