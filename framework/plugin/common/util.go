package common

import (
	"fmt"
	"reflect"

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
}

func NewPluginMgr(typeName string) *PluginMgr {
	ptr := &PluginMgr{
		typeName:   typeName,
		entityList: []any{},
	}
	return ptr
}

func (s *PluginMgr) getWeight(ptr any) (int, error) {
	vVal := reflect.ValueOf(ptr)
	funcVal := vVal.MethodByName(weightTag)
	if !funcVal.IsValid() {
		return DefaultWeight, nil
	}

	defer func() {
		if info := recover(); info != nil {
			slog.Error("panic in getWeight", "recover", info)
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

	return int(values[0].Int()), nil
}

func (s *PluginMgr) getID(ptr any) (string, error) {
	vVal := reflect.ValueOf(ptr)
	funcVal := vVal.MethodByName(idTag)
	if !funcVal.IsValid() {
		return "", fmt.Errorf("method %s not found", idTag)
	}

	defer func() {
		if info := recover(); info != nil {
			slog.Error("panic in getID", "recover", info)
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

	return values[0].String(), nil
}

func (s *PluginMgr) validPlugin(ptr any) error {
	vType := reflect.TypeOf(ptr)
	if vType.Kind() != reflect.Ptr {
		return fmt.Errorf("must be a pointer")
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

	newList := []any{}
	if len(s.entityList) == 0 {
		newList = append(newList, ptr)
	} else {
		ok := false
		for idx, val := range s.entityList {
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
	for _, val := range s.entityList {
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
	for _, val := range s.entityList {
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
	for _, val := range s.entityList {
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
	totalSize := len(s.entityList)
	for idx := range s.entityList {
		val := s.entityList[totalSize-idx-1]
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
