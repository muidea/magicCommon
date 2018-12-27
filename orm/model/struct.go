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
func GetStructInfo(objPtr interface{}) (ret *StructInfo, depends []*StructInfo, err error) {
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

func getStructInfo(structVal reflect.Value) (ret *StructInfo, depends []*StructInfo, err error) {
	ret = &StructInfo{name: structVal.Type().Name(), pkgPath: structVal.Type().PkgPath(), fields: make(Fields, 0)}
	depends = []*StructInfo{}

	structType := structVal.Type()
	fieldNum := structVal.NumField()

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

		ret.fields.Append(fieldInfo)

		fType := fieldInfo.GetFieldType()
		if !util.IsBasicType(fType.Value()) {
			fValue := fieldInfo.GetFieldValue()
			dvs, err := fValue.GetDepend()
			if err != nil {
				break
			}

			reference = append(reference, dvs...)
		}

		idx++
	}

	ret.primaryKey = ret.fields.GetPrimaryKey()

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

	return
}

func getStructPrimaryKey(structVal reflect.Value) (ret *FieldInfo, err error) {
	if structVal.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal value type, not struct, type:%s", structVal.Type().String())
		return
	}

	structType := structVal.Type()
	fieldNum := structVal.NumField()
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
