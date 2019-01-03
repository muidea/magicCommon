package orm

import (
	"log"
	"reflect"

	"muidea.com/magicCommon/orm/builder"
	"muidea.com/magicCommon/orm/model"
)

func (s *orm) insertSingle(structInfo model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	sql, err := builder.BuildInsert()
	if err != nil {
		return err
	}

	id := s.executor.Insert(sql)
	pk := structInfo.GetPrimaryField()
	if pk != nil {
		pk.SetFieldValue(reflect.ValueOf(id))
	}

	return
}

func (s *orm) insertRelation(structInfo model.StructInfo, fieldName string, relationInfo model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	relationSQL, relationErr := builder.BuildInsertRelation(fieldName, relationInfo)
	if relationErr != nil {
		err = relationErr
		return err
	}

	s.executor.Insert(relationSQL)
	return
}

func (s *orm) Insert(obj interface{}) (err error) {
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

	err = s.insertSingle(structInfo)
	if err != nil {
		return
	}

	for key, val := range structInfo.GetDependStructs() {
		if !val.IsStructPtr() {
			err = s.insertSingle(val)
			if err != nil {
				return
			}
		}

		err = s.insertRelation(structInfo, key, val)
		if err != nil {
			return
		}
	}

	return
}
