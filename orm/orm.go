package orm

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicCommon/orm/builder"
	"muidea.com/magicCommon/orm/executor"
	"muidea.com/magicCommon/orm/model"
)

// Orm orm interfalce
type Orm interface {
	Insert(obj interface{}) error
	Update(obj interface{}) error
	Delete(obj interface{}) error
	Query(obj interface{}, filter ...string) error
}

var ormManager *manager

func init() {
	ormManager = newManager()
}

type orm struct {
	executor executor.Executor
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

	executor, err := executor.NewExecutor(cfg.user, cfg.password, cfg.address, cfg.dbName)
	if err != nil {
		return nil, err
	}

	return &orm{executor: executor}, nil
}

func (s *orm) Insert(obj interface{}) error {
	modelInfo := model.GetStructInfo(obj)
	if modelInfo == nil {
		return fmt.Errorf("illegal model object, [%v]", obj)
	}

	builder := builder.NewBuilder(obj)
	_, err := ormManager.findModule(modelInfo.GetStructName())
	if err != nil {
		// no exist
		sql, err := builder.BuildCreateSchema()
		if err != nil {
			return err
		}

		err = ormManager.registerModule(modelInfo.GetStructName(), modelInfo.GetPkgPath())
		if err != nil {
			log.Printf("registerModule failed, err:%s", err.Error())
		}
		log.Print(sql)
	}

	sql, err := builder.BuildInsert()
	if err != nil {
		return err
	}

	id := s.executor.Insert(sql)
	pk := modelInfo.GetPrimaryKey()
	if pk != nil {
		pk.SetFieldValue(reflect.ValueOf(id))
	}

	log.Print(sql)

	return nil
}

func (s *orm) Update(obj interface{}) error {
	modelInfo := model.GetStructInfo(obj)
	if modelInfo == nil {
		return fmt.Errorf("illegal model object, [%v]", obj)
	}

	builder := builder.NewBuilder(obj)
	_, err := ormManager.findModule(modelInfo.GetStructName())
	if err != nil {
		// no exist
		sql, err := builder.BuildCreateSchema()
		if err != nil {
			return err
		}

		err = ormManager.registerModule(modelInfo.GetStructName(), modelInfo.GetPkgPath())
		if err != nil {
			log.Printf("registerModule failed, err:%s", err.Error())
		}
		log.Print(sql)
	}

	sql, err := builder.BuildUpdate()
	if err != nil {
		return err
	}
	log.Print(sql)

	return nil
}

func (s *orm) Delete(obj interface{}) error {
	modelInfo := model.GetStructInfo(obj)
	if modelInfo == nil {
		return fmt.Errorf("illegal model object, [%v]", obj)
	}

	builder := builder.NewBuilder(obj)
	_, err := ormManager.findModule(modelInfo.GetStructName())
	if err != nil {
		// no exist
		sql, err := builder.BuildCreateSchema()
		if err != nil {
			return err
		}

		err = ormManager.registerModule(modelInfo.GetStructName(), modelInfo.GetPkgPath())
		if err != nil {
			log.Printf("registerModule failed, err:%s", err.Error())
		}
		log.Print(sql)
	}

	sql, err := builder.BuildDelete()
	if err != nil {
		return err
	}
	log.Print(sql)

	return nil
}

func (s *orm) Query(obj interface{}, filter ...string) error {
	modelInfo := model.GetStructInfo(obj)
	if modelInfo == nil {
		return fmt.Errorf("illegal model object, [%v]", obj)
	}

	builder := builder.NewBuilder(obj)
	_, err := ormManager.findModule(modelInfo.GetStructName())
	if err != nil {
		// no exist
		sql, err := builder.BuildCreateSchema()
		if err != nil {
			return err
		}

		err = ormManager.registerModule(modelInfo.GetStructName(), modelInfo.GetPkgPath())
		if err != nil {
			log.Printf("registerModule failed, err:%s", err.Error())
		}
		log.Print(sql)
	}

	sql, err := builder.BuildQuery()
	if err != nil {
		return err
	}
	log.Print(sql)

	return nil
}
