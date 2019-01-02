package mysql

import (
	"fmt"
	"log"

	"muidea.com/magicCommon/orm/model"
)

// BuildDropSchema  BuildDropSchema
func (s *Builder) BuildDropSchema() (string, error) {
	str := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", s.getTableName(s.structInfo))
	log.Print(str)

	return str, nil
}

// BuildDropRelationSchema BuildDropRelationSchema
func (s *Builder) BuildDropRelationSchema(fieldName string, relationInfo model.StructInfo) (string, error) {
	str := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", s.GetRelationTableName(fieldName, relationInfo))
	log.Print(str)

	return str, nil
}
