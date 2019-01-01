package model

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicCommon/orm/util"
)

// FieldInfo FieldInfo
type FieldInfo interface {
	GetFieldName() string
	GetFieldType() FieldType
	GetFieldTag() FieldTag
	GetFieldValue() FieldValue
	SetFieldValue(val reflect.Value)
	Dump() string
}

// fieldInfo single field info
type fieldInfo struct {
	fieldIndex int
	fieldName  string

	fieldType  FieldType
	fieldTag   FieldTag
	fieldValue FieldValue
}

// Fields field info collection
type Fields []FieldInfo

// GetFieldName GetFieldName
func (s *fieldInfo) GetFieldName() string {
	return s.fieldName
}

// GetFieldType GetFieldType
func (s *fieldInfo) GetFieldType() FieldType {
	return s.fieldType
}

// GetFieldTag GetFieldTag
func (s *fieldInfo) GetFieldTag() FieldTag {
	return s.fieldTag
}

// GetFieldValue GetFieldValue
func (s *fieldInfo) GetFieldValue() FieldValue {
	return s.fieldValue
}

// SetFieldValue SetFieldValue
func (s *fieldInfo) SetFieldValue(val reflect.Value) {
	s.fieldValue.SetValue(val)
}

// Verify Verify
func (s *fieldInfo) Verify() error {
	if s.fieldTag.Name() == "" {
		return fmt.Errorf("no define field tag")
	}

	if s.fieldTag.IsAutoIncrement() {
		switch s.fieldType.Value() {
		case util.TypeBooleanField, util.TypeStringField, util.TypeDateTimeField, util.TypeFloatField, util.TypeDoubleField, util.TypeStructField, util.TypeSliceField:
			return fmt.Errorf("illegal auto_increment field type, type:%s", s.fieldType)
		default:
		}
	}

	if s.fieldTag.IsPrimaryKey() {
		switch s.fieldType.Value() {
		case util.TypeStructField, util.TypeSliceField:
			return fmt.Errorf("illegal primary key field type, type:%s", s.fieldType)
		default:
		}
	}

	return nil
}

// Dump Dump
func (s *fieldInfo) Dump() string {
	valStr, _ := s.fieldValue.GetValueStr()
	return fmt.Sprintf("index:[%d],name:[%s],type:[%s],tag:[%s],value:[%s]", s.fieldIndex, s.fieldName, s.fieldType, s.fieldTag, valStr)
}

// Append Append
func (s *Fields) Append(fieldInfo FieldInfo) {
	exist := false
	newField := fieldInfo.GetFieldTag()
	for _, val := range *s {
		curField := val.GetFieldTag()
		if curField.Name() == newField.Name() {
			exist = true
			break
		}
	}
	if exist {
		log.Fatalf("duplicate field tag,[%s]", fieldInfo.Dump())
	}

	*s = append(*s, fieldInfo)
}

// GetPrimaryKey get primarykey field
func (s *Fields) GetPrimaryKey() FieldInfo {
	for _, val := range *s {
		fieldTag := val.GetFieldTag()
		if fieldTag.IsPrimaryKey() {
			return val
		}
	}

	return nil
}

// Dump Dump
func (s *Fields) Dump() {
	for _, v := range *s {
		fmt.Printf("\t%s\n", v.Dump())
	}
}

// GetFieldInfo GetFieldInfo
func GetFieldInfo(idx int, fieldType reflect.StructField, fieldVal reflect.Value) (ret FieldInfo, err error) {
	info := &fieldInfo{}
	info.fieldIndex = idx
	info.fieldName = fieldType.Name

	info.fieldType, err = newFieldType(fieldType)
	if err != nil {
		return
	}

	info.fieldTag, err = newFieldTag(fieldType.Tag.Get("orm"))
	if err != nil {
		return
	}

	info.fieldValue, err = newFieldValue(fieldVal.Addr())
	if err != nil {
		return
	}

	err = info.Verify()
	if err != nil {
		return
	}

	ret = info
	return
}
