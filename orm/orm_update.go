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
	structInfo, structErr := model.GetObjectStructInfo(obj, s.modelInfoCache)
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

	fields := structInfo.GetDependField()
	for _, val := range fields {
		fType := val.GetFieldType()
		fDepend := fType.Depend()
		if fDepend == nil {
			continue
		}

		fValue := val.GetFieldValue()
		if fValue == nil {
			continue
		}
		fDependValue, fDependErr := fValue.GetDepend()
		if fDependErr != nil {
			err = fDependErr
			return
		}

		for _, fVal := range fDependValue {
			infoVal, infoErr := model.GetStructValue(fVal, s.modelInfoCache)
			if infoErr != nil {
				err = infoErr
				return
			}

			if !fType.IsPtr() {
				err = s.updateSingle(infoVal)
				if err != nil {
					return
				}
			}

			err = s.updateRelation(structInfo, val.GetFieldName(), infoVal)
			if err != nil {
				return
			}
		}
	}

	return
}
