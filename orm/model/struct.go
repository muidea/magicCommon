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

	primaryKey *FieldInfo

	fields Fields
}

// Verify Verify
func (s *StructInfo) Verify() error {
	if s.name == "" {
		return fmt.Errorf("illegal struct name")
	}

	return s.fields.Verify()
}

// GetStructName GetStructName
func (s *StructInfo) GetStructName() string {
	return s.name
}

// GetPkgPath GetPkgPath
func (s *StructInfo) GetPkgPath() string {
	return s.pkgPath
}

// GetFields GetFields
func (s *StructInfo) GetFields() *Fields {
	return &s.fields
}

// GetPrimaryKey GetPrimaryKey
func (s *StructInfo) GetPrimaryKey() *FieldInfo {
	return s.primaryKey
}

// Dump Dump
func (s *StructInfo) Dump() {
	fmt.Printf("name:%s, pkgPath:%s\n", s.name, s.pkgPath)
	if s.primaryKey != nil {
		fmt.Printf("primaryKey:%s", s.primaryKey.Dump())
	}
	s.fields.Dump()
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
	info := &StructInfo{name: val.Type().String(), pkgPath: val.Type().PkgPath(), fields: make(Fields, 0)}

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
			info.fields.Append(fInfo)
		} else {
			return nil
		}

		if fInfo.IsPrimaryKey() {
			info.primaryKey = fInfo
		}

		idx++
	}

	return info
}
