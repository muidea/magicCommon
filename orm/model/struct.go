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
	GetPrimaryField() FieldInfo
	GetDepends() map[string]StructInfo
	IsStructPtr() bool
	Dump()
}

// structInfo single struct ret
type structInfo struct {
	name    string
	pkgPath string

	fields Fields

	primaryKey FieldInfo

	depends map[string]StructInfo

	isStructPtr bool
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

func (s *structInfo) IsStructPtr() bool {
	return s.isStructPtr
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

// GetPrimaryField GetPrimaryField
func (s *structInfo) GetPrimaryField() FieldInfo {
	return s.primaryKey
}

func (s *structInfo) GetDepends() map[string]StructInfo {
	return s.depends
}

// Dump Dump
func (s *structInfo) Dump() {
	fmt.Print("structInfo:\n")
	fmt.Printf("\tname:%s, pkgPath:%s, isStructPtr:%v\n", s.name, s.pkgPath, s.isStructPtr)
	if s.primaryKey != nil {
		fmt.Printf("primaryKey:\n")
		fmt.Printf("\t%s\n", s.primaryKey.Dump())
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
	if structVal.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal obj type. must be a struct ptr")
		return
	}

	ret, err = GetStructInfoWithValue(structVal, cache)
	return
}

// GetStructInfoWithValue GetStructInfoWithValue
func GetStructInfoWithValue(structVal reflect.Value, cache StructInfoCache) (ret StructInfo, err error) {
	isStructPtr := false
	if structVal.Kind() == reflect.Ptr {
		structVal = reflect.Indirect(structVal)

		isStructPtr = true
	}

	structType := structVal.Type()
	if structType.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal structType, type:%s", structType.String())
		return
	}

	info := cache.Fetch(structType.Name())
	if info != nil {
		ret = info
		return
	}

	structInfo := &structInfo{name: structType.Name(), pkgPath: structType.PkgPath(), fields: make(Fields, 0), depends: map[string]StructInfo{}, isStructPtr: isStructPtr}

	fieldNum := structType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := structType.Field(idx)
		val := structVal.Field(idx)
		fieldValue := &val

		fieldInfo, fieldErr := GetFieldInfo(idx, fieldType, fieldValue)
		if fieldErr != nil {
			err = fieldErr
			log.Printf("getFieldInfo failed, name:%s, err:%s", fieldType.Name, err.Error())
			return
		}
		structInfo.fields.Append(fieldInfo)
	}

	structInfo.primaryKey = structInfo.fields.GetPrimaryField()

	cache.Put(structInfo.GetName(), structInfo)

	fields := structInfo.GetFields()
	for idx := range *fields {
		fieldInfo := (*fields)[idx]
		fType := fieldInfo.GetFieldType()
		fDepend := fType.Depend()
		if fDepend != nil {
			dStructInfo, dErr := GetStructInfo(fDepend, cache)
			if dErr != nil {
				err = dErr

				cache.Remove(structInfo.GetName())
				return
			}

			structInfo.depends[fieldInfo.GetFieldName()] = dStructInfo
		}
	}

	err = structInfo.Verify()
	if err != nil {
		return
	}

	ret = structInfo

	return
}

// GetStructInfo GetStructInfo
func GetStructInfo(structType reflect.Type, cache StructInfoCache) (ret StructInfo, err error) {
	isStructPtr := false
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
		isStructPtr = true
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

	structInfo := &structInfo{name: structType.Name(), pkgPath: structType.PkgPath(), fields: make(Fields, 0), depends: map[string]StructInfo{}, isStructPtr: isStructPtr}

	fieldNum := structType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := structType.Field(idx)
		fieldInfo, fieldErr := GetFieldInfo(idx, fieldType, nil)
		if fieldErr != nil {
			err = fieldErr
			log.Printf("getFieldInfo failed, name:%s, err:%s", fieldType.Name, err.Error())
			return
		}

		structInfo.fields.Append(fieldInfo)
	}

	structInfo.primaryKey = structInfo.fields.GetPrimaryField()

	cache.Put(structInfo.GetName(), structInfo)

	fields := structInfo.GetFields()
	for idx := range *fields {
		fieldInfo := (*fields)[idx]
		fType := fieldInfo.GetFieldType()
		fDepend := fType.Depend()
		if fDepend != nil {
			dStructInfo, dErr := GetStructInfo(fDepend, cache)
			if dErr != nil {
				err = dErr

				cache.Remove(structInfo.GetName())
				return
			}

			structInfo.depends[fieldInfo.GetFieldName()] = dStructInfo
		}
	}

	err = structInfo.Verify()
	if err != nil {
		cache.Remove(structInfo.GetName())
		return
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
