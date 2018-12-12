package mysql

import (
	"fmt"
	"log"

	"muidea.com/magicCommon/orm/model"
)

// Builder Builder
type Builder struct {
	obj interface{}
}

// New create builder
func New(obj interface{}) *Builder {
	return &Builder{obj: obj}
}

// BuildSchema  BuildSchema
func (s *Builder) BuildSchema() (string, error) {
	info := model.GetStructInfo(s.obj)
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
	for _, val := range info.Fields.Fields {
		if str == "" {
			str = fmt.Sprintf("%s %s", val.GetFieldTag(), val.GetFieldType())
		} else {
			str = fmt.Sprintf("%s, %s %s", str, s.getFieldName(val), s.getFieldType(val))
		}
	}
	if info.Fields.PrimaryKey != nil {
		str = fmt.Sprintf("%s, primary key ('%s')", str, s.getFieldName(info.Fields.PrimaryKey))
	}

	str = fmt.Sprintf("create table %s (%s)", s.getTableName(info), str)
	log.Print(str)

	return str, nil
}

func (s *Builder) getTableName(info *model.StructInfo) string {
	return info.GetStructName()
}

func (s *Builder) getFieldName(info *model.FieldInfo) string {
	return info.GetFieldTag()
}

func (s *Builder) getFieldType(info *model.FieldInfo) string {
	return info.GetFieldType()
}
