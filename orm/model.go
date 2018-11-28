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

	log.Print(reflect.Indirect(objVal).Type().Name())

	info := &modelInfo{name: reflect.Indirect(objVal).Type().Name(), fields: &fields{fields: make(map[string]*fieldInfo)}}

	fieldElem := objVal.Elem()
	fieldType := fieldElem.Type()
	idx := 0
	fieldNum := fieldElem.NumField()
	for {
		if idx >= fieldNum {
			break
		}

		sf := fieldElem.Field(idx)
		sv := fieldType.Field(idx)
		fInfo := getFieldInfo(idx, &sv, &sf)
		if fInfo != nil {
			info.fields.append(fInfo)
		} else {
			return nil
		}

		idx++
	}

	return info
}
