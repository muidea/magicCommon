package common

import (
	"fmt"
	"reflect"
	"sync"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/foundation/system"
	"github.com/muidea/magicCommon/task"
	"log/slog"
)

type InvokeFunc func() *cd.Error
type PluginMgr struct {
	typeName   string
	entityList []any
	mu         sync.RWMutex
}

func NewPluginMgr(typeName string) *PluginMgr {
	ptr := &PluginMgr{
		typeName:   typeName,
		entityList: []any{},
	}
	return ptr
}

func (s *PluginMgr) getWeight(ptr any) (weight int, err error) {
	vVal := reflect.ValueOf(ptr)
	funcVal := vVal.MethodByName(weightTag)
	if !funcVal.IsValid() {
		return DefaultWeight, nil
	}

	defer func() {
		if info := recover(); info != nil {
			slog.Error("panic in getWeight", "recover", info)
			weight = DefaultWeight
			err = fmt.Errorf("panic invoking %s", weightTag)
		}
	}()

	param := make([]reflect.Value, 0)
	values := funcVal.Call(param)
	if len(values) == 0 {
		return DefaultWeight, nil
	}

	if funcVal.Type().Out(0).String() != "int" {
		return DefaultWeight, fmt.Errorf("weight method must return int, got %s", funcVal.Type().Out(0).String())
	}

	weight = int(values[0].Int())
	return weight, nil
}

func (s *PluginMgr) getID(ptr any) (id string, err error) {
	vVal := reflect.ValueOf(ptr)
	funcVal := vVal.MethodByName(idTag)
	if !funcVal.IsValid() {
		return "", fmt.Errorf("method %s not found", idTag)
	}

	defer func() {
		if info := recover(); info != nil {
			slog.Error("panic in getID", "recover", info)
			id = ""
			err = fmt.Errorf("panic invoking %s", idTag)
		}
	}()

	param := make([]reflect.Value, 0)
	values := funcVal.Call(param)
	if len(values) == 0 {
		return "", fmt.Errorf("method %s returned no value", idTag)
	}

	if funcVal.Type().Out(0).String() != "string" {
		return "", fmt.Errorf("method %s must return string, got %s", idTag, funcVal.Type().Out(0).String())
	}

	id = values[0].String()
	return id, nil
}

func (s *PluginMgr) validPlugin(ptr any) error {
	if ptr == nil {
		return fmt.Errorf("plugin is nil")
	}

	vType := reflect.TypeOf(ptr)
	if vType == nil {
		return fmt.Errorf("plugin type is nil")
	}
	if vType.Kind() != reflect.Ptr {
		return fmt.Errorf("must be a pointer")
	}
	if reflect.ValueOf(ptr).IsNil() {
		return fmt.Errorf("plugin pointer is nil")
	}

	_, idOK := vType.MethodByName(idTag)
	//_, setupOK := vType.MethodByName(setupTag)
	_, runOK := vType.MethodByName(runTag)
	//_, teardownOK := vType.MethodByName(teardownTag)
	if !idOK || !runOK {
		return fmt.Errorf("invalid plugin value: missing required methods (ID, Run)")
	}

	return nil
}

func (s *PluginMgr) Register(ptr any) error {
	if err := s.validPlugin(ptr); err != nil {
		return err
	}

	curWeight, err := s.getWeight(ptr)
	if err != nil {
		return err
	}
	curID, err := s.getID(ptr)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	newList := []any{}
	if len(s.entityList) == 0 {
		newList = append(newList, ptr)
	} else {
		ok := false
		for idx, val := range s.entityList {
			existID, err := s.getID(val)
			if err != nil {
				return err
			}
			if existID == curID {
				return fmt.Errorf("duplicate plugin id: %s", curID)
			}

			preWeight, err := s.getWeight(val)
			if err != nil {
				return err
			}
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
	return nil
}

func (s *PluginMgr) GetEntity(id string) (ret any, err *cd.Error) {
	s.mu.RLock()
	entityList := append([]any(nil), s.entityList...)
	s.mu.RUnlock()

	for _, val := range entityList {
		idVal, idErr := s.getID(val)
		if idErr != nil {
			err = cd.NewError(cd.Unexpected, fmt.Sprintf("get ID failed: %v", idErr))
			return
		}
		if idVal == id {
			ret = val
			return
		}
	}

	err = cd.NewError(cd.Unexpected, fmt.Sprintf("%s not found", s.typeName))
	return
}

func (s *PluginMgr) Setup(eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) (err *cd.Error) {
	s.mu.RLock()
	entityList := append([]any(nil), s.entityList...)
	s.mu.RUnlock()

	for _, val := range entityList {
		err = system.InvokeEntityFunc(val, setupTag, eventHub, backgroundRoutine)
		if err != nil && err.Code != cd.NotFound {
			idVal, idErr := s.getID(val)
			if idErr != nil {
				slog.Error("invoke setup failed, get ID error", "type", s.typeName, "error", idErr)
			} else {
				slog.Error("invoke setup failed", "type", s.typeName, "id", idVal, "error", err)
			}
			return
		}
	}

	return
}

func (s *PluginMgr) Run() (err *cd.Error) {
	s.mu.RLock()
	entityList := append([]any(nil), s.entityList...)
	s.mu.RUnlock()

	for _, val := range entityList {
		err = system.InvokeEntityFunc(val, runTag)
		if err != nil && err.Code != cd.NotFound {
			idVal, idErr := s.getID(val)
			if idErr != nil {
				slog.Error("invoke run failed, get ID error", "type", s.typeName, "error", idErr)
			} else {
				slog.Error("invoke run failed", "type", s.typeName, "id", idVal, "error", err)
			}
			return
		}

		//slog.Info("invoke run success", "type", s.typeName, "id", idVal)
	}

	return
}

func (s *PluginMgr) Teardown() {
	s.mu.RLock()
	entityList := append([]any(nil), s.entityList...)
	s.mu.RUnlock()

	totalSize := len(entityList)
	for idx := range entityList {
		val := entityList[totalSize-idx-1]
		err := system.InvokeEntityFunc(val, teardownTag)
		if err != nil && err.Code != cd.NotFound {
			idVal, idErr := s.getID(val)
			if idErr != nil {
				slog.Error("invoke teardown failed, get ID error", "type", s.typeName, "error", idErr)
			} else {
				slog.Error("invoke teardown failed", "type", s.typeName, "id", idVal, "error", err)
			}
		}

		//slog.Info("invoke teardown success", "type", s.typeName, "id", idVal)
	}
}
