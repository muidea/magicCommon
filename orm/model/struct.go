package model

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicCommon/orm/util"
)

// StructInfo single struct ret
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

// UpdateFieldValue UpdateFieldValue
func (s *StructInfo) UpdateFieldValue(name string, val reflect.Value) error {
	for _, field := range s.fields {
		if field.GetFieldName() == name {
			field.SetFieldValue(val)
			return nil
		}
	}

	return fmt.Errorf("no found field, name:%s", name)
}

// GetPrimaryKey GetPrimaryKey
func (s *StructInfo) GetPrimaryKey() *FieldInfo {
	return s.primaryKey
}

// Dump Dump
func (s *StructInfo) Dump() {
	fmt.Print("structInfo:\n")
	fmt.Printf("\tname:%s, pkgPath:%s\n", s.name, s.pkgPath)
	if s.primaryKey != nil {
		fmt.Printf("primaryKey:\n")
		fmt.Printf("\t%s\n", s.primaryKey.Dump())
	}
	fmt.Print("fields:\n")
	s.fields.Dump()
}

// GetStructInfo GetStructInfo
func GetStructInfo(objPtr interface{}) (ret *StructInfo, depends []*StructInfo) {
	objType := reflect.TypeOf(objPtr)
	objVal := reflect.ValueOf(objPtr)

	if objType.Kind() != reflect.Ptr {
		log.Fatal("illegal struct type. must be a struct ptr")
	}

	structObj := reflect.Indirect(objVal)
	ret, depends = getStructInfo(structObj)
	return
}

func getStructInfo(structObj reflect.Value) (ret *StructInfo, depends []*StructInfo) {
	ret = &StructInfo{name: structObj.Type().String(), pkgPath: structObj.Type().PkgPath(), fields: make(Fields, 0)}
	depends = []*StructInfo{}

	structType := structObj.Type()
	fieldNum := structObj.NumField()

	idx := 0
	reference := []reflect.Value{}
	for {
		if idx >= fieldNum {
			break
		}

		fieldVal := structObj.Field(idx)
		fieldType := structType.Field(idx)
		fieldInfo := GetFieldInfo(idx, &fieldType, &fieldVal)
		if fieldInfo != nil {
			ret.fields.Append(fieldInfo)
		} else {
			return nil, nil
		}

		if fieldInfo.IsPrimaryKey() {
			ret.primaryKey = fieldInfo
		}

		if fieldInfo.GetFieldTypeValue() == util.TypeStructField {
			reference = append(reference, fieldInfo.GetFieldValue())
		}

		idx++
	}

	if len(reference) == 0 {
		return
	}

	for _, val := range reference {
		preRet, preDepends := getStructInfo(val)

		depends = append(preDepends, depends...)
		depends = append(depends, preRet)
	}

	return
}

func getStructValue(structObj reflect.Value) reflect.Value {
	ret, _ := getStructInfo(reflect.Indirect(structObj))

	pk := ret.GetPrimaryKey()
	if pk == nil {
		panic("illegal value, struct primary key is null")
	}

	if pk.GetFieldTypeValue() != util.TypeStructField {
		return reflect.Indirect(pk.GetFieldValue())
	}

	return getStructValue(pk.GetFieldValue())
}
