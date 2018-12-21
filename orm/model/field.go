package model

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicCommon/orm/util"
)

// FieldInfo single field info
type FieldInfo struct {
	fieldIndex int
	fieldName  string

	fieldType  FieldType
	fieldTag   FieldTag
	fieldValue FieldValue
}

// Fields field info collection
type Fields []*FieldInfo

// GetFieldName GetFieldName
func (s *FieldInfo) GetFieldName() string {
	return s.fieldName
}

// GetValueTypeEnum GetValueTypeEnum
func (s *FieldInfo) GetValueTypeEnum() FieldType {
	return s.fieldType
}

// GetFieldTag GetFieldTag
func (s *FieldInfo) GetFieldTag() FieldTag {
	return s.fieldTag
}

// GetFieldValue GetFieldValue
func (s *FieldInfo) GetFieldValue() FieldValue {
	return s.fieldValue
}

// SetFieldValue SetFieldValue
func (s *FieldInfo) SetFieldValue(val reflect.Value) {
	val = reflect.Indirect(val)
	s.fieldValue.SetValue(util.GetValue(s.fieldType.Value(), val))
}

// Verify Verify
func (s *FieldInfo) Verify() error {
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
func (s *FieldInfo) Dump() string {
	valStr, _ := s.fieldValue.GetValueStr()
	return fmt.Sprintf("index:%d,name:%s,type:%s,tag:%s,value:%s", s.fieldIndex, s.fieldName, s.fieldType, s.fieldTag, valStr)
}

// Append Append
func (s *Fields) Append(fieldType *FieldInfo) {
	exist := false
	for _, val := range *s {
		if val.fieldTag.Name() == fieldType.fieldTag.Name() {
			exist = true
			break
		}
	}
	if exist {
		log.Fatalf("duplicate field tag,[%s]", fieldType.Dump())
	}

	*s = append(*s, fieldType)
}

// GetPrimaryKey get primarykey field
func (s *Fields) GetPrimaryKey() *FieldInfo {
	for _, val := range *s {
		fieldTag := val.GetFieldTag()
		if fieldTag.IsPrimaryKey() {
			return val
		}
	}

	return nil
}

// Verify Verify
func (s *Fields) Verify() error {
	if len(*s) == 0 {
		return fmt.Errorf("no fields defined")
	}

	for _, val := range *s {
		err := val.Verify()
		if err != nil {
			return err
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
func GetFieldInfo(idx int, fieldType *reflect.StructField, fieldVal *reflect.Value) *FieldInfo {
	info := &FieldInfo{}
	info.fieldIndex = idx
	info.fieldName = fieldType.Name

	info.fieldType = newFieldType(fieldType)
	info.fieldTag = newFieldTag(fieldType.Tag.Get("orm"))
	info.fieldValue = newFieldValue(*fieldVal)

	return info
}
