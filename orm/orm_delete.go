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

func (s *orm) deleteRelation(structInfo, relationInfo model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	sql, err := builder.BuildDeleteRelation(relationInfo)
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
	structInfo, structDepends, structErr := model.GetObjectStructInfo(obj)
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

	for _, val := range structDepends {
		if !val.IsValuePtr() {
			err = s.deleteSingle(val)
			if err != nil {
				return
			}
		}

		err = s.deleteRelation(structInfo, val)
		if err != nil {
			return
		}
	}

	return
}
