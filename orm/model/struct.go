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
	GetDependStructs() (map[string]StructInfo, error)
	GetDependValues() (map[string][]reflect.Value, error)
	IsStructPtr() bool
	Dump()
}

// structInfo single struct ret
type structInfo struct {
	name    string
	pkgPath string

	fields Fields

	primaryKey FieldInfo

	isStructPtr bool

	structInfoCache StructInfoCache
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

// SetFieldValue SetFieldValue
func (s *structInfo) SetFieldValue(idx int, val reflect.Value) (err error) {
	for _, field := range s.fields {
		if field.GetFieldIndex() == idx {
			err = field.SetFieldValue(val)
			return
		}
	}

	err = fmt.Errorf("no found field, idx:%d", idx)
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
	return s.primaryKey
}

func (s *structInfo) GetDependStructs() (ret map[string]StructInfo, err error) {
	ret = map[string]StructInfo{}

	for _, field := range s.fields {
		fType := field.GetFieldType()
		fDepend := fType.Depend()
		if fDepend != nil {
			dStructInfo, dErr := GetStructInfo(fDepend, s.structInfoCache)
			if dErr != nil {
				err = dErr
				return
			}

			ret[field.GetFieldName()] = dStructInfo
		}
	}

	return
}

func (s *structInfo) GetDependValues() (ret map[string][]reflect.Value, err error) {
	ret = map[string][]reflect.Value{}

	for _, field := range s.fields {
		fValue := field.GetFieldValue()
		if fValue == nil {
			continue
		}

		fType := field.GetFieldType()
		fDepend := fType.Depend()
		if fDepend != nil {
			dVals, dErr := fValue.GetDepend()
			if dErr != nil {
				err = dErr
				return
			}

			ret[field.GetFieldName()] = dVals
		}
	}

	return
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

	structInfo := &structInfo{name: structType.Name(), pkgPath: structType.PkgPath(), fields: make(Fields, 0), isStructPtr: isStructPtr, structInfoCache: cache}

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

	structInfo := &structInfo{name: structType.Name(), pkgPath: structType.PkgPath(), fields: make(Fields, 0), isStructPtr: isStructPtr, structInfoCache: cache}

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

	ret = structInfo

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
		err = fmt.Errorf("can't get value from nil ptr")
		return
	}

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
