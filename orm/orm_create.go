package orm

import (
	"log"

	"muidea.com/magicCommon/orm/builder"
	"muidea.com/magicCommon/orm/model"
)

func (s *orm) createSchema(structInfo model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	tableName := builder.GetTableName()

	if !s.executor.CheckTableExist(tableName) {
		// no exist
		sql, err := builder.BuildCreateSchema()
		if err != nil {
			log.Printf("build create schema failed, err:%s", err.Error())
			return err
		}

		s.executor.Execute(sql)
	}

	return
}

func (s *orm) createRelationSchema(structInfo model.StructInfo, fieldName string, relationInfo model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	tableName := builder.GetRelationTableName(fieldName, relationInfo)

	if !s.executor.CheckTableExist(tableName) {
		// no exist
		sql, err := builder.BuildCreateRelationSchema(fieldName, relationInfo)
		if err != nil {
			return err
		}

		s.executor.Execute(sql)
	}

	return
}

func (s *orm) batchCreateSchema(structInfo model.StructInfo) (err error) {
	err = s.createSchema(structInfo)
	if err != nil {
		return
	}

	for key, val := range structInfo.GetDepends() {
		err = s.createSchema(val)
		if err != nil {
			return
		}

		err = s.createRelationSchema(structInfo, key, val)
		if err != nil {
			return
		}
	}

	return
}

func (s *orm) Create(obj interface{}) (err error) {
	structInfo, structErr := model.GetObjectStructInfo(obj, s.modelInfoCache)
	if structErr != nil {
		err = structErr
		log.Printf("GetObjectStructInfo failed, err:%s", err.Error())
		return
	}

	err = s.batchCreateSchema(structInfo)
	if err != nil {
		return
	}
	return
}
