package orm

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicCommon/orm/builder"
	"muidea.com/magicCommon/orm/model"
	"muidea.com/magicCommon/orm/util"
)

func (s *orm) querySingle(structInfo model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	sql, err := builder.BuildQuery()
	if err != nil {
		return err
	}

	s.executor.Query(sql)
	if !s.executor.Next() {
		return fmt.Errorf("no found object")
	}
	defer s.executor.Finish()

	items := []interface{}{}
	fields := structInfo.GetFields()
	for _, val := range *fields {
		fType := val.GetFieldType()
		if !util.IsBasicType(fType.Value()) {
			continue
		}

		v := util.GetBasicTypeInitValue(fType.Value())
		items = append(items, v)
	}
	s.executor.GetField(items...)

	idx := 0
	for _, val := range *fields {
		fType := val.GetFieldType()
		if !util.IsBasicType(fType.Value()) {
			continue
		}

		v := items[idx]
		val.SetFieldValue(reflect.Indirect(reflect.ValueOf(v)))
		idx++
	}

	return
}

func (s *orm) queryRelation(structInfo model.StructInfo, fieldName string, relationInfo model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	relationSQL, relationErr := builder.BuildQueryRelation(fieldName, relationInfo)
	if relationErr != nil {
		err = relationErr
		return err
	}

	s.executor.Query(relationSQL)
	return
}

func (s *orm) Query(obj interface{}, filter ...string) (err error) {
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

	//allStructInfos := structDepends
	//allStructInfos = append(allStructInfos, structInfo)
	//err = s.batchCreateSchema(allStructInfos)
	//if err != nil {
	//	return err
	//}

	err = s.querySingle(structInfo)

	return
}
