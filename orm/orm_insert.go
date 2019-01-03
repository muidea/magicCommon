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
		log.Printf("batchCreateSchema failed, err:%s", err.Error())
		return
	}

	err = s.insertSingle(structInfo)
	if err != nil {
		log.Printf("insertSingle failed, name:%s, err:%s", structInfo.GetName(), err.Error())
		return
	}

	dependVals, dependErr := structInfo.GetDependValues()
	if dependErr != nil {
		err = dependErr
		log.Printf("GetDependValues failed, name:%s, err:%s", structInfo.GetName(), err.Error())
		return
	}

	for key, val := range dependVals {
		for _, sv := range val {
			sInfo, sErr := model.GetStructValue(sv, s.modelInfoCache)
			if sErr != nil {
				err = sErr
				log.Printf("GetStructValue failed, err:%s", err.Error())
				return
			}

			if !sInfo.IsStructPtr() {
				err = s.insertSingle(sInfo)
				if err != nil {
					log.Printf("insertSingle failed, name:%s, err:%s", sInfo.GetName(), err.Error())
					return
				}
			}

			err = s.insertRelation(structInfo, key, sInfo)
			if err != nil {
				log.Printf("insertRelation failed, name:%s, fieldName:%s, relationName:%s, err:%s", structInfo.GetName(), key, sInfo.GetName(), err.Error())
				return
			}
		}
	}

	return
}
