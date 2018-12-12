package model

import (
	"fmt"
	"log"
	"reflect"
)

// StructInfo single struct info
type StructInfo struct {
	name    string
	pkgPath string
	Fields  *Fields
}

// Verify Verify
func (s *StructInfo) Verify() error {
	if s.name == "" {
		return fmt.Errorf("illegal struct name")
	}

	return s.Fields.Verify()
}

// GetStructName GetStructName
func (s *StructInfo) GetStructName() string {
	return s.name
}

// Dump Dump
func (s *StructInfo) Dump() {
	fmt.Printf("name:%s, pkgPath:%s\n", s.name, s.pkgPath)
	s.Fields.Dump()
}

// GetStructInfo GetStructInfo
func GetStructInfo(obj interface{}) *StructInfo {
	objType := reflect.TypeOf(obj)
	objVal := reflect.ValueOf(obj)

	if objType.Kind() != reflect.Ptr {
		log.Fatal("illegal struct value.")
		return nil
	}

	val := reflect.Indirect(objVal)
	info := &StructInfo{name: val.Type().String(), pkgPath: val.Type().PkgPath(), Fields: &Fields{Fields: make(map[string]*FieldInfo)}}

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
