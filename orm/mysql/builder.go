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

	str = fmt.Sprintf("CREATE TABLE `%s` (\n%s\n)\n", s.GetTableName(), str)
	log.Print(str)

	return str, nil
}

// BuildDropSchema  BuildDropSchema
func (s *Builder) BuildDropSchema() (string, error) {
	str := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", s.GetTableName())
	log.Print(str)

	return str, nil
}

// BuildInsert  BuildInsert
func (s *Builder) BuildInsert() (ret string, err error) {
	sql := ""
	vals, verr := s.getFieldValues(s.structInfo)
	if verr == nil {
		for _, val := range vals {
			sql = fmt.Sprintf("%sINSERT INTO `%s` (%s) VALUES (%s);", sql, s.GetTableName(), s.getFieldNames(s.structInfo, false), val)
		}
		log.Print(sql)
		ret = sql
	}
	err = verr

	return
}

// BuildUpdate  BuildUpdate
func (s *Builder) BuildUpdate() (ret string, err error) {
	str := ""
	for _, val := range *s.structInfo.GetFields() {
		fValue := val.GetFieldValue()
		fTag := val.GetFieldTag()
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
		str = fmt.Sprintf("UPDATE `%s` SET %s WHERE `%s`=%s", s.GetTableName(), str, pkfTag.Name(), pkfStr)
		log.Print(str)
	}

	ret = str
	err = pkferr

	return
}

// BuildDelete  BuildDelete
func (s *Builder) BuildDelete() (ret string, err error) {
	pkfValue := s.structInfo.GetPrimaryKey().GetFieldValue()
	pkfTag := s.structInfo.GetPrimaryKey().GetFieldTag()
	pkfStr, pkferr := pkfValue.GetValueStr()
	if pkferr == nil {
		ret = fmt.Sprintf("DELETE FROM `%s` WHERE `%s`=%s", s.GetTableName(), pkfTag.Name(), pkfStr)
		log.Print(ret)
	}

	err = pkferr

	return
}

// BuildQuery BuildQuery
func (s *Builder) BuildQuery() (ret string, err error) {
	pkfValue := s.structInfo.GetPrimaryKey().GetFieldValue()
	pkfTag := s.structInfo.GetPrimaryKey().GetFieldTag()
	pkfStr, pkferr := pkfValue.GetValueStr()
	if pkferr == nil {
		ret = fmt.Sprintf("SELECT %s FROM `%s` WHERE `%s`=%s", s.getFieldNames(s.structInfo, true), s.GetTableName(), pkfTag.Name(), pkfStr)
		log.Print(ret)
	}
	err = pkferr

	return
}

// GetTableName GetTableName
func (s *Builder) GetTableName() string {
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

func (s *Builder) getFieldValues(info *model.StructInfo) (ret []string, err error) {
	str := ""
	for _, field := range *info.GetFields() {
		fTag := field.GetFieldTag()
		if fTag.IsAutoIncrement() {
			continue
		}

		fType := field.GetFieldType()
		fValue := field.GetFieldValue()
		if fType.IsReference() {
			fValue, err = model.GetStructValue(fValue.GetValue())
		}

		if err != nil {
			break
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
