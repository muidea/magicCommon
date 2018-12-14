package mysql

import (
	"fmt"
	"log"
	"strings"

	"muidea.com/magicCommon/orm/model"
)

// Builder Builder
type Builder struct {
	obj             interface{}
	structInfoCache model.StructInfoCache
}

// New create builder
func New(obj interface{}, cache model.StructInfoCache) *Builder {
	return &Builder{obj: obj, structInfoCache: cache}
}

// BuildCreateSchema  BuildCreateSchema
func (s *Builder) BuildCreateSchema() (string, error) {
	info := model.GetStructInfo(s.obj, s.structInfoCache)
	if info == nil {
		return "", fmt.Errorf("get structInfo failed")
	}

	err := info.Verify()
	if err != nil {
		return "", err
	}

	err = verifyStructInfo(info)
	if err != nil {
		return "", err
	}

	str := ""
	for _, val := range *info.GetFields() {
		if str == "" {
			str = fmt.Sprintf("\t%s", declareFieldInfo(val))
		} else {
			str = fmt.Sprintf("%s,\n\t%s", str, declareFieldInfo(val))
		}
	}
	if info.GetPrimaryKey() != nil {
		str = fmt.Sprintf("%s,\n\tPRIMARY KEY (`%s`)", str, (info.GetPrimaryKey().GetFieldTag()))
	}

	str = fmt.Sprintf("CREATE TABLE `%s` (\n%s\n)\n", s.getTableName(info), str)
	log.Print(str)

	return str, nil
}

// BuildDropSchema  BuildDropSchema
func (s *Builder) BuildDropSchema() (string, error) {
	info := model.GetStructInfo(s.obj, s.structInfoCache)
	if info == nil {
		return "", fmt.Errorf("get structInfo failed")
	}

	err := info.Verify()
	if err != nil {
		return "", err
	}

	err = verifyStructInfo(info)
	if err != nil {
		return "", err
	}

	str := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", s.getTableName(info))
	log.Print(str)

	return str, nil
}

// BuildInsert  BuildInsert
func (s *Builder) BuildInsert() (string, error) {
	info := model.GetStructInfo(s.obj, s.structInfoCache)
	if info == nil {
		return "", fmt.Errorf("get structInfo failed")
	}

	err := info.Verify()
	if err != nil {
		return "", err
	}

	err = verifyStructInfo(info)
	if err != nil {
		return "", err
	}

	str := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", s.getTableName(info), s.getFieldNames(info, false), s.getFieldValues(info))
	log.Print(str)

	return str, nil
}

// BuildUpdate  BuildUpdate
func (s *Builder) BuildUpdate() (string, error) {
	info := model.GetStructInfo(s.obj, s.structInfoCache)
	if info == nil {
		return "", fmt.Errorf("get structInfo failed")
	}

	err := info.Verify()
	if err != nil {
		return "", err
	}

	err = verifyStructInfo(info)
	if err != nil {
		return "", err
	}

	str := ""
	for _, val := range *info.GetFields() {
		if val != info.GetPrimaryKey() {
			if str == "" {
				str = fmt.Sprintf("`%s`=%s", val.GetFieldTag(), val.GetFieldValueStr())
			} else {
				str = fmt.Sprintf("%s,`%s`=%s", str, val.GetFieldTag(), val.GetFieldValueStr())
			}
		}
	}

	str = fmt.Sprintf("UPDATE `%s` SET %s WHERE `%s`=%s", s.getTableName(info), str, info.GetPrimaryKey().GetFieldTag(), info.GetPrimaryKey().GetFieldValueStr())
	log.Print(str)

	return str, nil
}

// BuildDelete  BuildDelete
func (s *Builder) BuildDelete() (string, error) {
	info := model.GetStructInfo(s.obj, s.structInfoCache)
	if info == nil {
		return "", fmt.Errorf("get structInfo failed")
	}

	err := info.Verify()
	if err != nil {
		return "", err
	}

	err = verifyStructInfo(info)
	if err != nil {
		return "", err
	}

	str := fmt.Sprintf("DELETE FROM `%s` WHERE `%s`=%s", s.getTableName(info), info.GetPrimaryKey().GetFieldTag(), info.GetPrimaryKey().GetFieldValueStr())
	log.Print(str)

	return str, nil
}

// BuildQuery BuildQuery
func (s *Builder) BuildQuery() (string, error) {
	info := model.GetStructInfo(s.obj, s.structInfoCache)
	if info == nil {
		return "", fmt.Errorf("get structInfo failed")
	}

	err := info.Verify()
	if err != nil {
		return "", err
	}

	err = verifyStructInfo(info)
	if err != nil {
		return "", err
	}

	str := fmt.Sprintf("SELECT %s FROM `%s` WHERE `%s`=%s", s.getFieldNames(info, true), s.getTableName(info), info.GetPrimaryKey().GetFieldTag(), info.GetPrimaryKey().GetFieldValueStr())
	log.Print(str)

	return str, nil
}

func (s *Builder) getTableName(info *model.StructInfo) string {
	return strings.Join(strings.Split(info.GetStructName(), "."), "_")
}

func (s *Builder) getFieldNames(info *model.StructInfo, all bool) string {
	str := ""
	for _, field := range *info.GetFields() {
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
	for _, field := range *info.GetFields() {
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
