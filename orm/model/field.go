package model

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicCommon/orm"
)

// FieldInfo single field info
type FieldInfo struct {
	fieldIndex     int
	fieldName      string
	fieldTypeValue int
	fieldTypeName  string
	fieldTag       FieldTag
	fieldValue     reflect.Value
	fieldPkgPath   string
}

// Fields field info collection
type Fields struct {
	PrimaryKey *FieldInfo

	// name->FieldInfo
	Fields map[string]*FieldInfo
}

// GetFieldTag GetFieldTag
func (s *FieldInfo) GetFieldTag() string {
	return s.fieldTag.Name()
}

// GetFieldName GetFieldName
func (s *FieldInfo) GetFieldName() string {
	return s.fieldName
}

// GetFieldType GetFieldType
func (s *FieldInfo) GetFieldType() string {
	return s.fieldTypeName
}

// IsPrimaryKey IsPrimaryKey
func (s *FieldInfo) IsPrimaryKey() bool {
	return s.fieldTag.IsPrimaryKey()
}

// IsReference IsReference
func (s *FieldInfo) IsReference() bool {
	return s.fieldTypeValue >= orm.TypeStrictField
}

// Dump Dump
func (s *FieldInfo) Dump() string {
	return fmt.Sprintf("index:%d,name:%s,typeValue:%d, typeName:%s,tag:%s, pkgPath:%s", s.fieldIndex, s.fieldName, s.fieldTypeValue, s.fieldTypeName, s.fieldTag, s.fieldPkgPath)
}

// Append Append
func (s *Fields) Append(sf *FieldInfo) {
	_, ok := s.Fields[sf.fieldName]
	if ok {
		log.Fatalf("duplicate field,[%s]", sf.Dump())
	}

	s.Fields[sf.fieldName] = sf
}

// Verify Verify
func (s *Fields) Verify() error {
	if s.PrimaryKey == nil {
		return fmt.Errorf("no defined primary key")
	}

	if len(s.Fields) == 0 {
		return fmt.Errorf("no defined Fields")
	}

	return nil
}

// Dump Dump
func (s *Fields) Dump() {
	if s.PrimaryKey != nil {
		fmt.Printf("PrimaryKey:[%s]\n", s.PrimaryKey.Dump())
	}

	for k, v := range s.Fields {
		fmt.Printf("key:%s, val:[%s]\n", k, v.Dump())
	}
}

// GetFieldInfo GetFieldInfo
func GetFieldInfo(idx int, sf *reflect.StructField, sv *reflect.Value) *FieldInfo {
	info := &FieldInfo{}
	info.fieldIndex = idx
	info.fieldName = sf.Name
	info.fieldTag = newFieldTag(sf.Tag.Get("orm"))

	val := reflect.Indirect(*sv)
	tVal, err := GetFieldType(val.Type())
	if err != nil {
		log.Printf("GetFieldType failed, idx:%d, name:%s, type:%s, err:%s", idx, sf.Name, sf.Type.Kind(), err.Error())
		return nil
	}

	info.fieldTypeValue = tVal
	info.fieldTypeName = val.Type().String()
	info.fieldValue = val
	info.fieldPkgPath = val.Type().PkgPath()

	return info
}
