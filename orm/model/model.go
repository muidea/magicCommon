package model

import (
	"fmt"
	"log"
	"reflect"
)

// ModelInfo single model info
type ModelInfo struct {
	name    string
	pkgPath string
	Fields  *Fields
}

// Verify Verify
func (s *ModelInfo) Verify() error {
	if s.name == "" {
		return fmt.Errorf("illegal model name")
	}

	return s.Fields.Verify()
}

// Dump Dump
func (s *ModelInfo) Dump() {
	fmt.Printf("name:%s, pkgPath:%s\n", s.name, s.pkgPath)
	s.Fields.Dump()
}

// GetModelInfo GetModelInfo
func GetModelInfo(obj interface{}) *ModelInfo {
	objType := reflect.TypeOf(obj)
	objVal := reflect.ValueOf(obj)

	if objType.Kind() != reflect.Ptr {
		log.Fatal("illegal model value.")
		return nil
	}

	val := reflect.Indirect(objVal)
	info := &ModelInfo{name: val.Type().String(), pkgPath: val.Type().PkgPath(), Fields: &Fields{Fields: make(map[string]*FieldInfo)}}

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
		fInfo := GetFieldInfo(idx, &sf, &sv)
		if fInfo != nil {
			info.Fields.Append(fInfo)
		} else {
			return nil
		}

		if fInfo.IsPrimaryKey() {
			info.Fields.PrimaryKey = fInfo
		}

		idx++
	}

	return info
}
