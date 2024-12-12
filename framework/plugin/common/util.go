package common

import (
	"fmt"
	"reflect"
	"sync"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/foundation/system"
	"github.com/muidea/magicCommon/task"
)

type PluginMgr struct {
	typeName   string
	entityList []interface{}
}

func NewPluginMgr(typeName string) *PluginMgr {
	ptr := &PluginMgr{
		typeName:   typeName,
		entityList: []interface{}{},
	}
	return ptr
}

func (s *PluginMgr) getWeight(ptr interface{}) int {
	vVal := reflect.ValueOf(ptr)
	funcVal := vVal.MethodByName(weightTag)
	if !funcVal.IsValid() {
		return DefaultWeight
	}

	defer func() {
		if info := recover(); info != nil {
			err := fmt.Errorf("invoke %s unexpect, %v", weightTag, info)
			panic(err)
		}
	}()

	param := make([]reflect.Value, 0)
	values := funcVal.Call(param)
	if len(values) == 0 {
		return DefaultWeight
	}

	if funcVal.Type().Out(0).String() != "int" {
		return DefaultWeight
	}

	return int(values[0].Int())
}

func (s *PluginMgr) getID(ptr interface{}) string {
	vVal := reflect.ValueOf(ptr)
	funcVal := vVal.MethodByName(idTag)
	defer func() {
		if info := recover(); info != nil {
			err := fmt.Errorf("invoke %s unexpect, %v", idTag, info)
			panic(err)
		}
	}()

	param := make([]reflect.Value, 0)
	values := funcVal.Call(param)
	if len(values) == 0 {
		err := fmt.Errorf("invoke %s unexpect, illegal result", idTag)
		panic(err)
	}

	if funcVal.Type().Out(0).String() != "string" {
		err := fmt.Errorf("invoke %s unexpect, illegal result", idTag)
		panic(err)
	}

	return values[0].String()
}

func (s *PluginMgr) validEntity(ptr interface{}) {
	vType := reflect.TypeOf(ptr)
	if vType.Kind() != reflect.Ptr {
		panic("must be a pointer")
	}

	_, idOK := vType.MethodByName(idTag)
	//_, setupOK := vType.MethodByName(common.SetupTag)
	//_, runOK := vType.MethodByName(common.RunTag)
	//_, teardownOK := vType.MethodByName(common.TeardownTag)
	if !idOK {
		panic("invalid entity ptr")
	}
}

func (s *PluginMgr) Register(ptr interface{}) {
	s.validEntity(ptr)

	curWeight := s.getWeight(ptr)
	newList := []interface{}{}
	if len(s.entityList) == 0 {
		newList = append(newList, ptr)
	} else {
		ok := false
		for idx, val := range s.entityList {
			preWeight := s.getWeight(val)
			if preWeight <= curWeight {
				newList = append(newList, val)
				continue
			}

			ok = true
			newList = append(newList, ptr)
			newList = append(newList, s.entityList[idx:]...)
			break
		}

		if !ok {
			newList = append(newList, ptr)
		}
	}

	s.entityList = newList
}

func (s *PluginMgr) GetEntity(id string) (ret interface{}, err *cd.Result) {
	for _, val := range s.entityList {
		idVal := s.getID(val)
		if idVal == id {
			ret = val
			return
		}
	}

	err = cd.NewError(cd.UnExpected, fmt.Sprintf("%s not found", s.typeName))
	return
}

func (s *PluginMgr) invoke(wg *sync.WaitGroup, funcPtr func()) {
	if wg != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			funcPtr()
		}()

		return
	}

	funcPtr()
}

func (s *PluginMgr) Setup(eventHub event.Hub, backgroundRoutine task.BackgroundRoutine, wg *sync.WaitGroup) {
	for _, val := range s.entityList {
		setUp := func() {
			err := system.InvokeEntityFunc(val, setupTag, eventHub, backgroundRoutine)
			if err != nil && err.ErrorCode != cd.NoExist {
				log.Errorf("invoke %s %s setup failed, %v", s.typeName, s.getID(val), err)
				return
			}

			log.Infof("invoke %s %s setup success", s.typeName, s.getID(val))
		}

		s.invoke(wg, setUp)
	}
}

func (s *PluginMgr) Run(wg *sync.WaitGroup) {
	for _, val := range s.entityList {
		run := func() {
			err := system.InvokeEntityFunc(val, runTag)
			if err != nil && err.ErrorCode != cd.NoExist {
				log.Errorf("invoke %s %s run failed, %v", s.typeName, s.getID(val), err)
				return
			}

			log.Infof("invoke %s %s run success", s.typeName, s.getID(val))
		}

		s.invoke(wg, run)
	}
}

func (s *PluginMgr) Teardown(wg *sync.WaitGroup) {
	totalSize := len(s.entityList)
	for idx := range s.entityList {
		val := s.entityList[totalSize-idx-1]
		teardown := func() {
			err := system.InvokeEntityFunc(val, teardownTag)
			if err != nil && err.ErrorCode != cd.NoExist {
				log.Errorf("invoke %s %s teardown failed, %v", s.typeName, s.getID(val), err)
				return
			}

			log.Infof("invoke %s %s teardown success", s.typeName, s.getID(val))
		}

		s.invoke(wg, teardown)
	}
}
