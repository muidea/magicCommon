package mysql

import (
	"fmt"
	"log"

	"muidea.com/magicCommon/orm/model"
	"muidea.com/magicCommon/orm/util"
)

// BuildCreateSchema  BuildCreateSchema
func (s *Builder) BuildCreateSchema() (string, error) {
	str := ""
	for _, val := range *s.structInfo.GetFields() {
		fType := val.GetFieldType()
		if !util.IsBasicType(fType.Value()) {
			continue
		}

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

// BuildCreateRelationSchema BuildCreateRelationSchema
func (s *Builder) BuildCreateRelationSchema(relationInfo *model.StructInfo) (string, error) {
	str := "\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`left` INT NOT NULL,\n\t`right` INT NOT NULL,\n\tPRIMARY KEY (`id`)"
	str = fmt.Sprintf("CREATE TABLE `%s` (\n%s\n)\n", s.GetRelationTableName(relationInfo), str)
	log.Print(str)

	return str, nil
}
