package mysql

import (
	"fmt"
	"log"
	"strings"

	"muidea.com/magicCommon/orm"
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
		typeName, _ := s.getFieldType(val)
		if str == "" {
			str = fmt.Sprintf("\t`%s` %s NOT NULL", val.GetFieldTag(), typeName)
		} else {
			str = fmt.Sprintf("%s,\n\t`%s` %s NOT NULL", str, val.GetFieldTag(), typeName)
		}
	}
	if info.Fields.PrimaryKey != nil {
		str = fmt.Sprintf("%s, \n\tPRIMARY KEY (`%s`)", str, s.getFieldName(info.Fields.PrimaryKey))
	}

	str = fmt.Sprintf("CREATE TABLE `%s` (\n%s\n)\n", s.getTableName(info), str)
	log.Print(str)

	return str, nil
}

func (s *Builder) getTableName(info *model.StructInfo) string {
	return strings.Join(strings.Split(info.GetStructName(), "."), "_")
}

func (s *Builder) getFieldName(info *model.FieldInfo) string {
	return info.GetFieldTag()
}

func (s *Builder) getFieldType(info *model.FieldInfo) (ret string, err error) {

	typeValue := info.GetFieldTypeValue()
	switch typeValue {
	case orm.TypeBooleanField:
		ret = "TINYINT"
		err = nil
		break
	case orm.TypeVarCharField:
		ret = "TEXT"
		err = nil
		break
	case orm.TypeDateTimeField:
		ret = "DATETIME"
		err = nil
		break
	case orm.TypeBitField:
		ret = "TINYINT"
		err = nil
		break
	case orm.TypeSmallIntegerField:
		ret = "SMALLINT"
		err = nil
		break
	case orm.TypeIntegerField:
		ret = "INT"
		err = nil
		break
	case orm.TypeBigIntegerField:
		ret = "BIGINT"
		err = nil
		break
	case orm.TypePositiveBitField:
		ret = "SMALLINT"
		err = nil
		break
	case orm.TypePositiveSmallIntegerField:
		ret = "INT"
		err = nil
		break
	case orm.TypePositiveIntegerField:
		ret = "BIGINT"
		err = nil
		break
	case orm.TypePositiveBigIntegerField:
		ret = "BIGINT"
		err = nil
		break
	case orm.TypeFloatField:
		ret = "FLOAT"
		err = nil
		break
	case orm.TypeDoubleField:
		ret = "DOUBLE"
		err = nil
		break
	case orm.TypeStrictField:
		break
	default:
		err = fmt.Errorf("no support fileType, %d", typeValue)
	}

	return
}
