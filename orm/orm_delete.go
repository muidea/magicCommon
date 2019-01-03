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

	err = s.deleteSingle(structInfo)
	if err != nil {
		return
	}

	dependVals, dependErr := structInfo.GetDependValues()
	if dependErr != nil {
		err = dependErr
		return
	}

	for key, val := range dependVals {
		for _, sv := range val {
			sInfo, sErr := model.GetStructValue(sv, s.modelInfoCache)
			if sErr != nil {
				err = sErr
				return
			}

			if !sInfo.IsStructPtr() {
				err = s.deleteSingle(sInfo)
				if err != nil {
					return
				}
			}

			err = s.deleteRelation(structInfo, key, sInfo)
			if err != nil {
				return
			}
		}
	}

	return
}
