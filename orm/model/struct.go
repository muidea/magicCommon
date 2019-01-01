package model

import (
	"fmt"
	"log"
	"reflect"
)

// StructInfo StructInfo
type StructInfo interface {
	GetName() string
	GetPkgPath() string
	GetFields() *Fields
	UpdateFieldValue(name string, val reflect.Value) error
	GetPrimaryKey() FieldInfo
	Dump()
}

// structInfo single struct ret
type structInfo struct {
	name    string
	pkgPath string

	primaryKey FieldInfo

	fields Fields
}

// Verify Verify
func (s *structInfo) Verify() error {
	if s.name == "" {
		return fmt.Errorf("illegal struct name")
	}

	return nil
}

func (s *structInfo) GetName() string {
	return s.name
}

// GetPkgPath GetPkgPath
func (s *structInfo) GetPkgPath() string {
	return s.pkgPath
}

// GetFields GetFields
func (s *structInfo) GetFields() *Fields {
	return &s.fields
}

// UpdateFieldValue UpdateFieldValue
func (s *structInfo) UpdateFieldValue(name string, val reflect.Value) error {
	for _, field := range s.fields {
		if field.GetFieldName() == name {
			field.SetFieldValue(val)
			return nil
		}
	}

	return fmt.Errorf("no found field, name:%s", name)
}

// GetPrimaryKey GetPrimaryKey
func (s *structInfo) GetPrimaryKey() FieldInfo {
	return s.primaryKey
}

// Dump Dump
func (s *structInfo) Dump() {
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
func GetStructInfo(objPtr interface{}) (ret StructInfo, depends []StructInfo, err error) {
	ptrVal := reflect.ValueOf(objPtr)

	if ptrVal.Kind() != reflect.Ptr {
		err = fmt.Errorf("illegal obj type. must be a struct ptr")
		return
	}

	structVal := reflect.Indirect(ptrVal)
	if structVal.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal obj type. must be a struct ptr")
		return
	}

	ret, depends, err = getStructInfo(structVal)
	return
}

func getStructInfo(structVal reflect.Value) (ret StructInfo, depends []StructInfo, err error) {
	structInfo := &structInfo{name: structVal.Type().Name(), pkgPath: structVal.Type().PkgPath(), fields: make(Fields, 0)}
	depends = []StructInfo{}

	structType := structVal.Type()
	fieldNum := structType.NumField()

	idx := 0
	reference := []reflect.Value{}
	for {
		if idx >= fieldNum {
			break
		}

		fieldType := structType.Field(idx)
		fieldVal := structVal.Field(idx)
		fieldInfo, fieldErr := GetFieldInfo(idx, fieldType, fieldVal)
		if fieldErr != nil {
			err = fieldErr
			log.Printf("getFieldInfo failed, err:%s", err.Error())
			return
		}

		structInfo.fields.Append(fieldInfo)

		fType := fieldInfo.GetFieldType()
		fDepend := fType.Depend()
		if fDepend != nil {
			fValue := fieldInfo.GetFieldValue()
			dvs, err := fValue.GetDepend()
			if err != nil {
				break
			}

			reference = append(reference, dvs...)
		}
		idx++
	}

	structInfo.primaryKey = structInfo.fields.GetPrimaryKey()

	err = structInfo.Verify()
	if err != nil {
		return
	}

	if len(reference) == 0 {
		return
	}

	for _, val := range reference {
		preRet, preDepends, err := getStructInfo(reflect.Indirect(val))
		if err != nil {
			break
		}

		depends = append(preDepends, depends...)
		depends = append(depends, preRet)
	}

	ret = structInfo

	return
}

func getStructPrimaryKey(structVal reflect.Value) (ret FieldInfo, err error) {
	if structVal.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal value type, not struct, type:%s", structVal.Type().String())
		return
	}

	structType := structVal.Type()
	fieldNum := structType.NumField()
	for idx := 0; idx < fieldNum; {
		fieldType := structType.Field(idx)
		fieldVal := structVal.Field(idx)
		fieldInfo, fieldErr := GetFieldInfo(idx, fieldType, fieldVal)
		if fieldErr != nil {
			err = fieldErr
			return
		}

		fTag := fieldInfo.GetFieldTag()
		if fTag.IsPrimaryKey() {
			ret = fieldInfo
			return
		}

		idx++
	}

	err = fmt.Errorf("no found primary key. type:%s", structVal.Type().String())
	return
}
