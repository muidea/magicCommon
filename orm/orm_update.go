package orm

import (
	"fmt"
	"log"

	"muidea.com/magicCommon/orm/builder"
	"muidea.com/magicCommon/orm/model"
)

func (s *orm) updateSingle(structInfo *model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	sql, err := builder.BuildUpdate()
	if err != nil {
		return err
	}

	num := s.executor.Update(sql)
	if num != 1 {
		log.Printf("unexception update, rowNum:%d", num)
		err = fmt.Errorf("update %s failed", structInfo.GetStructName())
	}

	return err
}

func (s *orm) Update(obj interface{}) (err error) {
	structInfo, structDepends, structErr := model.GetStructInfo(obj)
	if structErr != nil {
		err = structErr
		return
	}

	err = s.batchCreateSchema(structInfo, structDepends)
	if err != nil {
		return
	}

	for _, val := range structDepends {
		err = s.updateSingle(val)
		if err != nil {
			return
		}
	}

	err = s.updateSingle(structInfo)

	return
}
