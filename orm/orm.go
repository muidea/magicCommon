package orm

import (
	"fmt"
	"log"

	"muidea.com/magicCommon/orm/builder"
	"muidea.com/magicCommon/orm/model"
	"muidea.com/magicCommon/orm/mysql"
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
	executor mysql.Executor
}

// New create new Orm
func New() Orm {
	return &orm{}
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
