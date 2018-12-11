package orm

import (
	"fmt"
	"log"
	"reflect"
)

// single model info
type modelInfo struct {
	name   string
	fields *fields
}

func (s *modelInfo) verify() error {
	if s.name == "" {
		return fmt.Errorf("illegal model name")
	}

	return s.fields.verify()
}

func (s *modelInfo) Dump() {
	fmt.Printf("name:%s\n", s.name)
	s.fields.Dump()
}

func getModelInfo(obj interface{}) *modelInfo {
	objType := reflect.TypeOf(obj)
	objVal := reflect.ValueOf(obj)

	if objType.Kind() != reflect.Ptr {
		log.Fatal("illegal model value.")
		return nil
	}

	info := &modelInfo{name: reflect.Indirect(objVal).Type().String(), fields: &fields{fields: make(map[string]*fieldInfo)}}

	fieldElem := objVal.Elem()
	fieldType := fieldElem.Type()
	idx := 0
	fieldNum := fieldElem.NumField()
	for {
		if idx >= fieldNum {
			break
		}

		sv := fieldElem.Field(idx)
		sf := fieldType.Field(idx)
		fInfo := getFieldInfo(idx, &sf, &sv)
		if fInfo != nil {
			info.fields.append(fInfo)
		} else {
			return nil
		}

		if fInfo.isPrimaryKey() {
			info.fields.pk = fInfo
		}

		idx++
	}

	return info
}
