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

		dependType, _ := fType.Depend()
		if dependType != nil {
			continue
		}

		v := util.GetBasicTypeInitValue(fType.Value())
		items = append(items, v)
	}
	s.executor.GetField(items...)

	idx := 0
	for _, val := range *fields {
		fType := val.GetFieldType()

		dependType, _ := fType.Depend()
		if dependType != nil {
			continue
		}

		v := items[idx]
		err = val.SetFieldValue(reflect.Indirect(reflect.ValueOf(v)))
		if err != nil {
			return err
		}

		idx++
	}

	return
}

func (s *orm) queryRelation(structInfo model.StructInfo, fieldInfo model.FieldInfo, relationInfo model.StructInfo) (err error) {
	fValue := fieldInfo.GetFieldValue()
	if fValue == nil || fValue.IsNil() {
		return
	}

	builder := builder.NewBuilder(structInfo)
	relationSQL, relationErr := builder.BuildQueryRelation(fieldInfo.GetFieldName(), relationInfo)
	if relationErr != nil {
		err = relationErr
		return err
	}

	fType := fieldInfo.GetFieldType()
	values := []int{}

	func() {
		s.executor.Query(relationSQL)
		defer s.executor.Finish()
		for s.executor.Next() {
			v := 0
			s.executor.GetField(&v)
			values = append(values, v)
		}
	}()

	if util.IsStructType(fType.Value()) {
		if len(values) > 0 {
			fDepend, _ := fType.Depend()
			relationVal := reflect.New(fDepend)
			relationInfo, relationErr = model.GetStructValue(relationVal, s.modelInfoCache)
			if relationErr != nil {
				err = relationErr
				return
			}

			relationInfo.GetPrimaryField().SetFieldValue(reflect.ValueOf(values[0]))
			err = s.querySingle(relationInfo)
			if err != nil {
				return
			}

			structInfo.UpdateFieldValue(fieldInfo.GetFieldName(), relationVal)
		}
	} else if util.IsSliceType(fType.Value()) {
		sizeLen := len(values)
		relationVal, _ := fValue.GetValue()
		relationType := relationVal.Type()
		if fType.IsPtr() {
			relationType = relationType.Elem()
		}

		relationVal = reflect.MakeSlice(relationType, sizeLen, sizeLen)
		for idx, val := range values {
			fDepend, fDependPtr := fType.Depend()
			itemVal := reflect.New(fDepend)
			itemInfo, itemErr := model.GetStructValue(itemVal, s.modelInfoCache)
			if itemErr != nil {
				log.Printf("GetStructValue faield, err:%s", itemErr.Error())
				err = itemErr
				return
			}

			itemInfo.GetPrimaryField().SetFieldValue(reflect.ValueOf(val))
			err = s.querySingle(itemInfo)
			if err != nil {
				return
			}

			if !fDependPtr {
				itemVal = reflect.Indirect(itemVal)
			}

			relationVal.Index(idx).Set(itemVal)
		}
		structInfo.UpdateFieldValue(fieldInfo.GetFieldName(), relationVal)
	}

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

	err = s.querySingle(structInfo)
	if err != nil {
		return
	}

	fields := structInfo.GetDependField()
	for _, val := range fields {
		fType := val.GetFieldType()
		fDepend, _ := fType.Depend()

		if fDepend == nil {
			continue
		}

		infoVal, infoErr := model.GetStructInfo(fDepend, s.modelInfoCache)
		if infoErr != nil {
			err = infoErr
			return
		}
		err = s.queryRelation(structInfo, val, infoVal)
		if err != nil {
			return
		}
	}

	return
}
