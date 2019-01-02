package orm

import (
	"fmt"
	"log"

	"muidea.com/magicCommon/orm/builder"
	"muidea.com/magicCommon/orm/model"
)

func (s *orm) updateSingle(structInfo model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	sql, err := builder.BuildUpdate()
	if err != nil {
		return err
	}

	num := s.executor.Update(sql)
	if num != 1 {
		log.Printf("unexception update, rowNum:%d", num)
		err = fmt.Errorf("update %s failed", structInfo.GetName())
	}

	return err
}

func (s *orm) updateRelation(structInfo model.StructInfo, fieldName string, relationInfo model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	relationSQL, relationErr := builder.BuildUpdateRelation(fieldName, relationInfo)
	if relationErr != nil {
		err = relationErr
		return err
	}

	s.executor.Update(relationSQL)
	return
}

func (s *orm) Update(obj interface{}) (err error) {
	structInfo, structErr := model.GetObjectStructInfo(obj, true, s.modelInfoCache)
	if structErr != nil {
		err = structErr
		log.Printf("GetObjectStructInfo failed, err:%s", err.Error())
		return
	}

	//err = s.batchCreateSchema(structInfo, structDepends)
	//if err != nil {
	//	return
	//}

	err = s.updateSingle(structInfo)
	if err != nil {
		return
	}

	for key, val := range structInfo.GetDepends() {
		if !val.IsStructPtr() {
			err = s.updateSingle(val)
			if err != nil {
				return
			}
		}

		err = s.updateRelation(structInfo, key, val)
		if err != nil {
			return
		}
	}

	return
}
