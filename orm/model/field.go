package model

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicCommon/orm/util"
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
type Fields []*FieldInfo

// GetFieldTag GetFieldTag
func (s *FieldInfo) GetFieldTag() string {
	return s.fieldTag.Name()
}

// GetFieldName GetFieldName
func (s *FieldInfo) GetFieldName() string {
	return s.fieldName
}

// GetFieldTypeName GetFieldTypeName
func (s *FieldInfo) GetFieldTypeName() string {
	return s.fieldTypeName
}

// GetFieldTypeValue GetFieldTypeValue
func (s *FieldInfo) GetFieldTypeValue() int {
	return s.fieldTypeValue
}

// SetFieldValue SetFieldValue
func (s *FieldInfo) SetFieldValue(val reflect.Value) {
	s.fieldValue.Set(val)
}

// GetFieldValue GetFieldValue
func (s *FieldInfo) GetFieldValue() reflect.Value {
	return s.fieldValue
}

// GetFieldValueStr GetFieldValueStr
func (s *FieldInfo) GetFieldValueStr() (ret string) {
	switch s.fieldTypeValue {
	case util.TypeBooleanField:
		if s.fieldValue.Bool() {
			ret = "1"
		} else {
			ret = "0"
		}
		break
	case util.TypeVarCharField:
		ret = fmt.Sprintf("'%s'", s.fieldValue.Interface())
		break
	case util.TypeDateTimeField:
		ret = fmt.Sprintf("'%s'", s.fieldValue.Interface())
		break
	case util.TypeBitField:
		ret = fmt.Sprintf("%d", s.fieldValue.Interface())
		break
	case util.TypeSmallIntegerField:
		ret = fmt.Sprintf("%d", s.fieldValue.Interface())
		break
	case util.TypeIntegerField:
		ret = fmt.Sprintf("%d", s.fieldValue.Interface())
		break
	case util.TypeBigIntegerField:
		ret = fmt.Sprintf("%d", s.fieldValue.Interface())
		break
	case util.TypePositiveBitField:
		ret = fmt.Sprintf("%d", s.fieldValue.Interface())
		break
	case util.TypePositiveSmallIntegerField:
		ret = fmt.Sprintf("%d", s.fieldValue.Interface())
		break
	case util.TypePositiveIntegerField:
		ret = fmt.Sprintf("%d", s.fieldValue.Interface())
		break
	case util.TypePositiveBigIntegerField:
		ret = fmt.Sprintf("%d", s.fieldValue.Interface())
		break
	case util.TypeFloatField:
		ret = fmt.Sprintf("%f", s.fieldValue.Interface())
		break
	case util.TypeDoubleField:
		ret = fmt.Sprintf("%f", s.fieldValue.Interface())
		break
	default:
		msg := fmt.Sprintf("no support fileType, %d", s.fieldTypeValue)
		panic(msg)
	}

	return
}

// IsPrimaryKey IsPrimaryKey
func (s *FieldInfo) IsPrimaryKey() bool {
	return s.fieldTag.IsPrimaryKey()
}

// IsAutoIncrement IsAutoIncrement
func (s *FieldInfo) IsAutoIncrement() bool {
	return s.fieldTag.IsAutoIncrement()
}

// IsReference IsReference
func (s *FieldInfo) IsReference() bool {
	return s.fieldTypeValue >= util.TypeStrictField
}

// Dump Dump
func (s *FieldInfo) Dump() string {
	return fmt.Sprintf("index:%d,name:%s,typeValue:%d, typeName:%s,tag:%s, pkgPath:%s", s.fieldIndex, s.fieldName, s.fieldTypeValue, s.fieldTypeName, s.fieldTag, s.fieldPkgPath)
}

// Append Append
func (s *Fields) Append(sf *FieldInfo) {
	exist := false
	for _, val := range *s {
		if val.fieldTag.Name() == sf.fieldTag.Name() {
			exist = true
			break
		}
	}
	if exist {
		log.Fatalf("duplicate field tag,[%s]", sf.Dump())
	}

	*s = append(*s, sf)
}

// Verify Verify
func (s *Fields) Verify() error {
	if len(*s) == 0 {
		return fmt.Errorf("no defined Fields")
	}

	return nil
}

// Dump Dump
func (s *Fields) Dump() {
	for k, v := range *s {
		fmt.Printf("key:%d, val:[%s]\n", k, v.Dump())
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
