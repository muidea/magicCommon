package model

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicCommon/orm/util"
)

// FieldInfo FieldInfo
type FieldInfo interface {
	GetFieldIndex() int
	GetFieldName() string
	GetFieldType() FieldType
	GetFieldTag() FieldTag
	GetFieldValue() FieldValue
	SetFieldValue(val reflect.Value) error
	Copy() FieldInfo
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

func (s *fieldInfo) GetFieldIndex() int {
	return s.fieldIndex
}

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
func (s *fieldInfo) SetFieldValue(val reflect.Value) (err error) {
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return
		}
	}

	if s.fieldValue != nil {
		err = s.fieldValue.SetValue(val)
	} else {
		s.fieldValue, err = newFieldValue(val.Addr())
	}

	return
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

func (s *fieldInfo) Copy() FieldInfo {
	var fieldValue FieldValue
	if s.fieldValue != nil {
		fieldValue = s.fieldValue.Copy()
	}

	return &fieldInfo{
		fieldIndex: s.fieldIndex,
		fieldName:  s.fieldName,
		fieldType:  s.fieldType.Copy(),
		fieldTag:   s.fieldTag.Copy(),
		fieldValue: fieldValue,
	}
}

// Dump Dump
func (s *fieldInfo) Dump() string {
	str := fmt.Sprintf("index:[%d],name:[%s],type:[%s],tag:[%s]", s.fieldIndex, s.fieldName, s.fieldType, s.fieldTag)
	if s.fieldValue != nil {
		valStr, _ := s.fieldValue.GetValueStr()

		str = fmt.Sprintf("%s,value:[%s]", str, valStr)
	}

	return str
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

// GetPrimaryField get primarykey field
func (s *Fields) GetPrimaryField() FieldInfo {
	for _, val := range *s {
		fieldTag := val.GetFieldTag()
		if fieldTag.IsPrimaryKey() {
			return val
		}
	}

	return nil
}

// Copy Copy
func (s *Fields) Copy() Fields {
	ret := make(Fields, 0)
	for _, val := range *s {
		ret = append(ret, val.Copy())
	}
	return ret
}

// Dump Dump
func (s *Fields) Dump() {
	for _, v := range *s {
		fmt.Printf("\t%s\n", v.Dump())
	}
}

// GetFieldInfo GetFieldInfo
func GetFieldInfo(idx int, fieldType reflect.StructField, fieldVal *reflect.Value) (ret FieldInfo, err error) {
	ormStr := fieldType.Tag.Get("orm")
	if ormStr == "" {
		return
	}

	info := &fieldInfo{}
	info.fieldIndex = idx
	info.fieldName = fieldType.Name

	info.fieldType, err = newFieldType(fieldType)
	if err != nil {
		return
	}

	info.fieldTag, err = newFieldTag(ormStr)
	if err != nil {
		return
	}

	if fieldVal != nil {
		info.fieldValue, err = newFieldValue(fieldVal.Addr())
		if err != nil {
			return
		}
	}

	err = info.Verify()
	if err != nil {
		return
	}

	ret = info
	return
}
