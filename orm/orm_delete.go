package orm

import (
	"fmt"
	"log"

	"muidea.com/magicCommon/orm/builder"
	"muidea.com/magicCommon/orm/model"
)

func (s *orm) deleteSingle(structInfo model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	sql, err := builder.BuildDelete()
	if err != nil {
		return err
	}
	num := s.executor.Delete(sql)
	if num != 1 {
		log.Printf("unexception delete, rowNum:%d", num)
		err = fmt.Errorf("delete %s failed", structInfo.GetName())
	}

	return
}

func (s *orm) deleteRelation(structInfo model.StructInfo, fieldName string, relationInfo model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	sql, err := builder.BuildDeleteRelation(fieldName, relationInfo)
	if err != nil {
		return err
	}
	num := s.executor.Delete(sql)
	if num != 1 {
		log.Printf("unexception delete, rowNum:%d", num)
		err = fmt.Errorf("delete %s relation failed", structInfo.GetName())
	}

	return
}

func (s *orm) Delete(obj interface{}) (err error) {
	structInfo, structErr := model.GetObjectStructInfo(obj, s.modelInfoCache)
	if structErr != nil {
		err = structErr
		log.Printf("GetObjectStructInfo failed, err:%s", err.Error())
		return
	}

	//err = s.batchCreateSchema(structInfo, structDepends)

	err = s.deleteSingle(structInfo)
	if err != nil {
		return
	}

	for key, val := range structInfo.GetDependStructs() {
		if !val.IsStructPtr() {
			err = s.deleteSingle(val)
			if err != nil {
				return
			}
		}

		err = s.deleteRelation(structInfo, key, val)
		if err != nil {
			return
		}
	}

	return
}
