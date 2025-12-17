package common

import (
	"fmt"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/foundation/system"
	"github.com/muidea/magicCommon/task"
)

type InvokeFunc func() *cd.Error
type PluginMgr struct {
	typeName   string
	entityList []any
}

func NewPluginMgr(typeName string) *PluginMgr {
	ptr := &PluginMgr{
		typeName:   typeName,
		entityList: []any{},
	}
	return ptr
}

func (s *PluginMgr) getWeight(ptr any) int {
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

func (s *PluginMgr) getID(ptr any) string {
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

func (s *PluginMgr) validPlugin(ptr any) {
	vType := reflect.TypeOf(ptr)
	if vType.Kind() != reflect.Ptr {
		panic("must be a pointer")
	}

	_, idOK := vType.MethodByName(idTag)
	//_, setupOK := vType.MethodByName(setupTag)
	_, runOK := vType.MethodByName(runTag)
	//_, teardownOK := vType.MethodByName(teardownTag)
	if !idOK || !runOK {
		panic("invalid plugin value")
	}
}

func (s *PluginMgr) Register(ptr any) {
	s.validPlugin(ptr)

	curWeight := s.getWeight(ptr)
	newList := []any{}
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

func (s *PluginMgr) GetEntity(id string) (ret any, err *cd.Error) {
	for _, val := range s.entityList {
		idVal := s.getID(val)
		if idVal == id {
			ret = val
			return
		}
	}

	err = cd.NewError(cd.Unexpected, fmt.Sprintf("%s not found", s.typeName))
	return
}

func (s *PluginMgr) Setup(eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) (err *cd.Error) {
	for _, val := range s.entityList {
		err = system.InvokeEntityFunc(val, setupTag, eventHub, backgroundRoutine)
		if err != nil && err.Code != cd.NotFound {
			log.Errorf("invoke [%s:%s]->setup failed, %v", s.typeName, s.getID(val), err)
			return
		}
	}

	return
}

func (s *PluginMgr) Run() (err *cd.Error) {
	for _, val := range s.entityList {
		err = system.InvokeEntityFunc(val, runTag)
		if err != nil && err.Code != cd.NotFound {
			log.Errorf("invoke [%s:%s]->run failed, %v", s.typeName, s.getID(val), err)
			return
		}

		//log.Infof("invoke %s %s run success", s.typeName, s.getID(val))
	}

	return
}

func (s *PluginMgr) Teardown() {
	totalSize := len(s.entityList)
	for idx := range s.entityList {
		val := s.entityList[totalSize-idx-1]
		err := system.InvokeEntityFunc(val, teardownTag)
		if err != nil && err.Code != cd.NotFound {
			log.Errorf("invoke [%s:%s]->teardown failed, %v", s.typeName, s.getID(val), err)
		}

		//log.Infof("invoke %s %s teardown success", s.typeName, s.getID(val))
	}
}
