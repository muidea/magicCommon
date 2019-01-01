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
	pk := structInfo.GetPrimaryKey()
	if pk != nil {
		pk.SetFieldValue(reflect.ValueOf(id))
	}

	return
}

func (s *orm) insertRelation(structInfo, relationInfo model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	relationSQL, relationErr := builder.BuildInsertRelation(relationInfo)
	if relationErr != nil {
		err = relationErr
		return err
	}

	s.executor.Insert(relationSQL)
	return
}

func (s *orm) Insert(obj interface{}) (err error) {
	structInfo, structDepends, structErr := model.GetStructInfo(obj)
	if structErr != nil {
		err = structErr
		log.Printf("getStructInfo failed, err:%s", err.Error())
		return
	}

	err = s.batchCreateSchema(structInfo, structDepends)
	if err != nil {
		return
	}

	err = s.insertSingle(structInfo)
	if err != nil {
		return
	}

	for _, val := range structDepends {
		err = s.insertSingle(val)
		if err != nil {
			return
		}
		err = s.insertRelation(structInfo, val)
		if err != nil {
			return
		}
	}

	return
}
