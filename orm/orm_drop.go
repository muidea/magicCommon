package orm

import (
	"log"

	"muidea.com/magicCommon/orm/builder"
	"muidea.com/magicCommon/orm/model"
)

func (s *orm) dropSingle(structInfo model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	tableName := builder.GetTableName()
	if s.executor.CheckTableExist(tableName) {
		sql, err := builder.BuildDropSchema()
		if err != nil {
			return err
		}

		s.executor.Execute(sql)
	}

	return
}

func (s *orm) dropRelation(structInfo model.StructInfo, fieldName string, relationInfo model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	tableName := builder.GetRelationTableName(fieldName, relationInfo)
	if s.executor.CheckTableExist(tableName) {
		sql, err := builder.BuildDropRelationSchema(fieldName, relationInfo)
		if err != nil {
			return err
		}

		s.executor.Execute(sql)
	}

	return
}

func (s *orm) Drop(obj interface{}) (err error) {
	structInfo, structErr := model.GetObjectStructInfo(obj, s.modelInfoCache)
	if structErr != nil {
		err = structErr
		log.Printf("GetObjectStructInfo failed, err:%s", err.Error())
		return
	}

	err = s.dropSingle(structInfo)
	if err != nil {
		return
	}

	dependVals, dependErr := structInfo.GetDependValues()
	if dependErr != nil {
		err = dependErr
		return
	}

	for key, val := range dependVals {
		for _, sv := range val {
			sInfo, sErr := model.GetStructValue(sv, s.modelInfoCache)
			if sErr != nil {
				err = sErr
				return
			}

			if !sInfo.IsStructPtr() {
				err = s.dropSingle(sInfo)
				if err != nil {
					return
				}
			}

			err = s.dropRelation(structInfo, key, sInfo)
			if err != nil {
				return
			}
		}
	}

	return
}
