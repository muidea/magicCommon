package model

import (
	"fmt"
	"reflect"
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

// IsSame IsSame
func (s *StructInfo) IsSame(info *StructInfo) bool {
	return s.name == info.name
}

// IsConflict IsConflict
func (s *StructInfo) IsConflict(info *StructInfo) bool {
	return s.name == info.name && s.pkgPath != info.pkgPath
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
	objType := reflect.TypeOf(objPtr)
	objVal := reflect.ValueOf(objPtr)

	if objType.Kind() != reflect.Ptr {
		err = fmt.Errorf("illegal obj type. must be a struct ptr")
		return
	}

	structObj := reflect.Indirect(objVal)
	if structObj.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal obj type. must be a struct ptr")
		return
	}
	ret, depends, err = getStructInfo(structObj)
	return
}

func getStructInfo(structObj reflect.Value) (ret *StructInfo, depends []*StructInfo, err error) {
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

		fieldType := structType.Field(idx)
		fieldVal := structObj.Field(idx)
		fieldInfo := GetFieldInfo(idx, &fieldType, &fieldVal)
		ret.fields.Append(fieldInfo)

		fv := fieldInfo.GetFieldValue()
		dvs, err := fv.GetDepend()
		if err != nil {
			break
		}

		reference = append(reference, dvs...)

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
