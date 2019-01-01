package orm

import (
	"log"

	"muidea.com/magicCommon/orm/builder"
	"muidea.com/magicCommon/orm/model"
)

func (s *orm) dropSingle(structInfo model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	tableName := builder.GetTableName()
	info := s.modelInfoCache.Fetch(tableName)
	if info != nil {
		sql, err := builder.BuildDropSchema()
		if err != nil {
			return err
		}

		s.executor.Execute(sql)
	}

	s.modelInfoCache.Remove(tableName)
	return
}

func (s *orm) dropRelation(structInfo, relationInfo model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	tableName := builder.GetRelationTableName(relationInfo)
	info := s.modelInfoCache.Fetch(tableName)
	if info != nil {
		sql, err := builder.BuildDropRelationSchema(relationInfo)
		if err != nil {
			return err
		}

		s.executor.Execute(sql)
	}

	s.modelInfoCache.Remove(tableName)
	return
}

func (s *orm) Drop(obj interface{}) (err error) {
	structInfo, structDepends, structErr := model.GetObjectStructInfo(obj)
	if structErr != nil {
		err = structErr
		log.Printf("GetObjectStructInfo failed, err:%s", err.Error())
		return
	}

	err = s.dropSingle(structInfo)
	if err != nil {
		return
	}

	for _, val := range structDepends {
		if !val.IsValuePtr() {
			err = s.dropSingle(val)
			if err != nil {
				return
			}
		}

		err = s.dropRelation(structInfo, val)
		if err != nil {
			return
		}
	}

	return
}
