package orm

import (
	"log"

	"muidea.com/magicCommon/orm/builder"
	"muidea.com/magicCommon/orm/model"
)

func (s *orm) createSchema(structInfo model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	tableName := builder.GetTableName()

	info := s.modelInfoCache.Fetch(tableName)
	if info == nil {
		if !s.executor.CheckTableExist(tableName) {
			// no exist
			sql, err := builder.BuildCreateSchema()
			if err != nil {
				log.Printf("build create schema failed, err:%s", err.Error())
				return err
			}

			s.executor.Execute(sql)
		}

		s.modelInfoCache.Put(tableName, structInfo)
	}

	return
}

func (s *orm) createRelationSchema(structInfo, relationInfo model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	tableName := builder.GetRelationTableName(relationInfo)

	info := s.modelInfoCache.Fetch(tableName)
	if info == nil {
		if !s.executor.CheckTableExist(tableName) {
			// no exist
			sql, err := builder.BuildCreateRelationSchema(relationInfo)
			if err != nil {
				return err
			}

			s.executor.Execute(sql)
		}

		s.modelInfoCache.Put(tableName, structInfo)
	}

	return
}

func (s *orm) batchCreateSchema(structInfo model.StructInfo, depends []model.StructInfo) (err error) {
	err = s.createSchema(structInfo)
	if err != nil {
		return
	}

	for _, val := range depends {
		err = s.createSchema(val)
		if err != nil {
			return
		}

		err = s.createRelationSchema(structInfo, val)
		if err != nil {
			return
		}
	}

	return
}
