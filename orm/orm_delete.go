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

func (s *orm) deleteRelation(structInfo model.StructInfo, fieldInfo model.FieldInfo) (err error) {
	fType := fieldInfo.GetFieldType()
	fDepend, fDependPtr := fType.Depend()

	if fDepend == nil {
		return
	}

	infoVal, infoErr := model.GetStructInfo(fDepend, s.modelInfoCache)
	if infoErr != nil {
		err = infoErr
		return
	}

	builder := builder.NewBuilder(structInfo)
	rightSQL, relationSQL, err := builder.BuildDeleteRelation(fieldInfo.GetFieldName(), infoVal)
	if err != nil {
		return err
	}

	if !fDependPtr {
		s.executor.Delete(rightSQL)
	}

	s.executor.Delete(relationSQL)

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

	fields := structInfo.GetDependField()
	for _, val := range fields {
		err = s.deleteRelation(structInfo, val)
		if err != nil {
			return
		}
	}

	return
}
