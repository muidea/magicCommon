package orm

import (
	"fmt"
	"reflect"

	"muidea.com/magicCommon/orm/builder"
	"muidea.com/magicCommon/orm/model"
	"muidea.com/magicCommon/orm/util"
)

func (s *orm) querySingle(structInfo *model.StructInfo) (err error) {
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
		v := util.GetInitValue(fType.Value())
		items = append(items, v)
	}
	s.executor.GetField(items...)

	idx := 0
	for _, val := range *fields {
		v := items[idx]
		val.SetFieldValue(reflect.Indirect(reflect.ValueOf(v)))
		idx++
	}

	return
}

func (s *orm) Query(obj interface{}, filter ...string) (err error) {
	structInfo, structDepends, structErr := model.GetStructInfo(obj)
	if structErr != nil {
		err = structErr
		return
	}

	err = s.batchCreateSchema(structInfo, structDepends)
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
