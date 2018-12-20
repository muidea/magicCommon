package orm

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicCommon/orm/builder"
	"muidea.com/magicCommon/orm/executor"
	"muidea.com/magicCommon/orm/model"
	"muidea.com/magicCommon/orm/util"
)

// Orm orm interfalce
type Orm interface {
	Insert(obj interface{}) error
	Update(obj interface{}) error
	Delete(obj interface{}) error
	Query(obj interface{}, filter ...string) error
	Drop(obj interface{}) error
	Release()
}

var ormManager *manager

func init() {
	ormManager = newManager()
}

type orm struct {
	executor       executor.Executor
	modelInfoCache model.StructInfoCache
}

// Initialize InitOrm
func Initialize(user, password, address, dbName string) error {
	cfg := &serverConfig{user: user, password: password, address: address, dbName: dbName}

	ormManager.updateServerConfig(cfg)

	return nil
}

// Uninitialize Uninitialize orm
func Uninitialize() {

}

// New create new Orm
func New() (Orm, error) {
	cfg := ormManager.getServerConfig()
	if cfg == nil {
		return nil, fmt.Errorf("not define databaes server config")
	}

	executor, err := executor.NewExecutor(cfg.user, cfg.password, cfg.address, cfg.dbName)
	if err != nil {
		return nil, err
	}

	return &orm{executor: executor, modelInfoCache: ormManager.getCache()}, nil
}

func (s *orm) createSchema(structInfo *model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	tableName := builder.GetTableName()

	info := s.modelInfoCache.Fetch(tableName)
	if info == nil {
		if !s.executor.CheckTableExist(tableName) {
			// no exist
			sql, err := builder.BuildCreateSchema()
			if err != nil {
				return err
			}

			s.executor.Execute(sql)
		}

		s.modelInfoCache.Put(tableName, structInfo)
	}

	return
}

func (s *orm) createRelationSchema(structInfo *model.StructInfo, relationInfo *model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	tableName := builder.GetRelationTableName(relationInfo)

	info := s.modelInfoCache.Fetch(tableName)
	if info == nil {
		if !s.executor.CheckTableExist(tableName) {
			// no exist
			sql, err := builder.BuildCreateRelationSchema(relationInfo)
			if err != nil {
				return err
			}

			s.executor.Execute(sql)
		}

		s.modelInfoCache.Put(tableName, structInfo)
	}

	return
}

func (s *orm) batchCreateSchema(structInfo *model.StructInfo, depends []*model.StructInfo) (err error) {
	for _, val := range depends {
		err = s.createSchema(val)
		if err != nil {
			return
		}

		err = s.createRelationSchema(structInfo, val)
		if err != nil {
			return
		}
	}

	err = s.createSchema(structInfo)

	return nil
}

func (s *orm) insertSingle(structInfo *model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	sql, err := builder.BuildInsert()
	if err != nil {
		return err
	}

	id := s.executor.Insert(sql)
	pk := structInfo.GetPrimaryKey()
	if pk != nil {
		pk.SetFieldValue(reflect.ValueOf(id))
	}

	return
}

func (s *orm) Insert(obj interface{}) (err error) {
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
		err = s.insertSingle(val)
		if err != nil {
			return
		}
	}

	err = s.insertSingle(structInfo)

	return
}

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

func (s *orm) deleteSingle(structInfo *model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	sql, err := builder.BuildDelete()
	if err != nil {
		return err
	}
	num := s.executor.Delete(sql)
	if num != 1 {
		log.Printf("unexception delete, rowNum:%d", num)
		err = fmt.Errorf("delete %s failed", structInfo.GetStructName())
	}

	return
}

func (s *orm) Delete(obj interface{}) (err error) {
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
		err = s.deleteSingle(val)
		if err != nil {
			return
		}
	}

	s.deleteSingle(structInfo)

	return
}

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

func (s *orm) dropSingle(structInfo *model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	tableName := builder.GetTableName()
	info := s.modelInfoCache.Fetch(tableName)
	if info != nil {
		sql, err := builder.BuildDropSchema()
		if err != nil {
			return err
		}

		s.executor.Execute(sql)
	}

	s.modelInfoCache.Remove(tableName)
	return
}

func (s *orm) dropRelation(structInfo *model.StructInfo, relationInfo *model.StructInfo) (err error) {
	builder := builder.NewBuilder(structInfo)
	tableName := builder.GetRelationTableName(relationInfo)
	info := s.modelInfoCache.Fetch(tableName)
	if info != nil {
		sql, err := builder.BuildDropRelationSchema(relationInfo)
		if err != nil {
			return err
		}

		s.executor.Execute(sql)
	}

	s.modelInfoCache.Remove(tableName)
	return
}

func (s *orm) Drop(obj interface{}) (err error) {
	structInfo, structDepends, structErr := model.GetStructInfo(obj)
	if structErr != nil {
		err = structErr
		return
	}

	for _, val := range structDepends {
		err = s.dropSingle(val)
		if err != nil {
			return
		}

		err = s.dropRelation(structInfo, val)
		if err != nil {
			return
		}
	}

	s.dropSingle(structInfo)

	return nil
}

func (s *orm) Release() {
	if s.executor != nil {
		s.executor.Release()
		s.executor = nil
	}
}
