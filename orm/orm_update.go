package orm

import (
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

	s.executor.Update(sql)

	return err
}

func (s *orm) updateRelation(structInfo model.StructInfo, fieldInfo model.FieldInfo) (err error) {
	err = s.deleteRelation(structInfo, fieldInfo)
	if err != nil {
		return
	}

	err = s.insertRelation(structInfo, fieldInfo)
	if err != nil {
		return
	}

	return
}

func (s *orm) Update(obj interface{}) (err error) {
	structInfo, structErr := model.GetObjectStructInfo(obj, s.modelInfoCache)
	if structErr != nil {
		err = structErr
		log.Printf("GetObjectStructInfo failed, err:%s", err.Error())
		return
	}

	err = s.updateSingle(structInfo)
	if err != nil {
		return
	}

	fields := structInfo.GetDependField()
	for _, val := range fields {
		err = s.updateRelation(structInfo, val)
		if err != nil {
			return
		}
	}

	return
}
