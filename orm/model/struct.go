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
	SetFieldValue(idx int, val reflect.Value) error
	UpdateFieldValue(name string, val reflect.Value) error
	GetPrimaryField() FieldInfo
	GetDependField() []FieldInfo
	Copy() StructInfo
	Dump()
}

// structInfo single struct ret
type structInfo struct {
	name    string
	pkgPath string

	fields Fields

	structInfoCache StructInfoCache
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

// SetFieldValue SetFieldValue
func (s *structInfo) SetFieldValue(idx int, val reflect.Value) (err error) {
	for _, field := range s.fields {
		if field.GetFieldIndex() == idx {
			err = field.SetFieldValue(val)
			return
		}
	}

	return
}

// UpdateFieldValue UpdateFieldValue
func (s *structInfo) UpdateFieldValue(name string, val reflect.Value) (err error) {
	for _, field := range s.fields {
		if field.GetFieldName() == name {
			err = field.SetFieldValue(val)
			return
		}
	}

	err = fmt.Errorf("no found field, name:%s", name)
	return
}

// GetPrimaryField GetPrimaryField
func (s *structInfo) GetPrimaryField() FieldInfo {
	return s.fields.GetPrimaryField()
}

func (s *structInfo) GetDependField() (ret []FieldInfo) {
	for _, field := range s.fields {
		fType := field.GetFieldType()
		fDepend, _ := fType.Depend()
		if fDepend != nil {
			ret = append(ret, field)
		}
	}

	return
}

func (s *structInfo) Copy() StructInfo {
	info := &structInfo{name: s.name, pkgPath: s.pkgPath, fields: s.fields.Copy(), structInfoCache: s.structInfoCache}
	return info
}

// Dump Dump
func (s *structInfo) Dump() {
	fmt.Print("structInfo:\n")
	fmt.Printf("\tname:%s, pkgPath:%s\n", s.name, s.pkgPath)

	primaryKey := s.fields.GetPrimaryField()
	if primaryKey != nil {
		fmt.Printf("primaryKey:\n")
		fmt.Printf("\t%s\n", primaryKey.Dump())
	}
	fmt.Print("fields:\n")
	s.fields.Dump()
}

// GetObjectStructInfo GetObjectStructInfo
func GetObjectStructInfo(objPtr interface{}, cache StructInfoCache) (ret StructInfo, err error) {
	ptrVal := reflect.ValueOf(objPtr)

	if ptrVal.Kind() != reflect.Ptr {
		err = fmt.Errorf("illegal obj type. must be a struct ptr")
		return
	}

	structVal := reflect.Indirect(ptrVal)
	structType := structVal.Type()

	ret, err = GetStructInfo(structType, cache)
	if err != nil {
		log.Printf("GetStructInfo failed, err:%s", err.Error())
		return
	}

	ret, err = GetStructValue(structVal, cache)
	if err != nil {
		log.Printf("GetStructValue failed, err:%s", err.Error())
		return
	}

	return
}

// GetStructInfo GetStructInfo
func GetStructInfo(structType reflect.Type, cache StructInfoCache) (ret StructInfo, err error) {
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	if structType.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal structType, type:%s", structType.String())
		return
	}

	info := cache.Fetch(structType.Name())
	if info != nil {
		ret = info
		return
	}

	structInfo := &structInfo{name: structType.Name(), pkgPath: structType.PkgPath(), fields: make(Fields, 0), structInfoCache: cache}

	fieldNum := structType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := structType.Field(idx)
		fieldInfo, fieldErr := GetFieldInfo(idx, fieldType, nil)
		if fieldErr != nil {
			err = fieldErr
			log.Printf("getFieldInfo failed, name:%s, err:%s", fieldType.Name, err.Error())
			return
		}

		if fieldInfo != nil {
			structInfo.fields.Append(fieldInfo)
		}
	}

	if len(structInfo.fields) > 0 {
		cache.Put(structInfo.GetName(), structInfo)

		ret = structInfo
		return
	}

	err = fmt.Errorf("no define orm field, struct name:%s", structInfo.GetName())
	return
}

// GetStructValue GetStructValue
func GetStructValue(structVal reflect.Value, cache StructInfoCache) (ret StructInfo, err error) {
	if structVal.Kind() == reflect.Ptr {
		if structVal.IsNil() {
			err = fmt.Errorf("can't get value from nil ptr")
			return
		}

		structVal = reflect.Indirect(structVal)
	}

	info := cache.Fetch(structVal.Type().Name())
	if info == nil {
		err = fmt.Errorf("can't get value structInfo, valType:%s", structVal.Type().String())
		return
	}

	info = info.Copy()
	fieldNum := structVal.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		val := structVal.Field(idx)
		err = info.SetFieldValue(idx, val)
		if err != nil {
			log.Printf("SetFieldValue failed, err:%s", err.Error())
			return
		}
	}

	ret = info

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
		fieldInfo, fieldErr := GetFieldInfo(idx, fieldType, &fieldVal)
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
