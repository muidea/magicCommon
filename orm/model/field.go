package model

import (
	"fmt"
	"log"
	"reflect"
	"time"

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
	val = reflect.Indirect(val)
	switch s.fieldTypeValue {
	case util.TypeBooleanField:
		if val.Int() > 0 {
			s.fieldValue.SetBool(true)
		} else {
			s.fieldValue.SetBool(false)
		}
	case util.TypeDoubleField, util.TypeFloatField:
		s.fieldValue.SetFloat(val.Float())
	case util.TypeDateTimeField:
		ts, _ := time.ParseInLocation("2006-01-02 15:04:05", val.String(), time.Local)
		s.fieldValue.Set(reflect.ValueOf(ts))
	case util.TypeBitField, util.TypeSmallIntegerField, util.TypeIntegerField, util.TypeInteger32Field, util.TypeBigIntegerField:
		s.fieldValue.SetInt(val.Int())
	case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveIntegerField, util.TypePositiveInteger32Field, util.TypePositiveBigIntegerField:
		s.fieldValue.SetUint(val.Uint())
	case util.TypeStringField:
		s.fieldValue.SetString(val.String())
	case util.TypeStructField:
		reallyVal := reflect.Indirect(s.fieldValue)
		reallyVal.Set(val)
	default:
		msg := fmt.Sprintf("unexception value, name:%s, pkgPath:%s, type:%s, valueType:%s", s.fieldName, s.fieldPkgPath, s.fieldTypeName, val.Kind())
		panic(msg)
	}
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
	case util.TypeStringField:
		ret = fmt.Sprintf("'%s'", s.fieldValue.Interface())
		break
	case util.TypeDateTimeField:
		ts, ok := s.fieldValue.Interface().(time.Time)
		if ok {
			ret = fmt.Sprintf("'%s'", ts.Format("2006-01-02 15:04:05"))
		} else {
			msg := fmt.Sprintf("illegal value,[%v]", s.fieldValue.Interface())
			panic(msg)
		}
		break
	case util.TypeBitField, util.TypePositiveBitField,
		util.TypeSmallIntegerField, util.TypePositiveSmallIntegerField,
		util.TypeIntegerField, util.TypePositiveIntegerField,
		util.TypeInteger32Field, util.TypePositiveInteger32Field,
		util.TypeBigIntegerField, util.TypePositiveBigIntegerField:
		ret = fmt.Sprintf("%d", s.fieldValue.Interface())
		break
	case util.TypeFloatField, util.TypeDoubleField:
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

// Verify Verify
func (s *FieldInfo) Verify() error {
	if s.fieldTag.Name() == "" {
		return fmt.Errorf("no define field tag")
	}

	if s.IsAutoIncrement() {
		switch s.fieldTypeValue {
		case util.TypeBooleanField, util.TypeStringField, util.TypeDateTimeField, util.TypeFloatField, util.TypeDoubleField, util.TypeStructField:
			return fmt.Errorf("illegal auto_increment field type, type:%s", s.fieldTypeName)
		default:
		}
	}

	if s.IsPrimaryKey() {
		switch s.fieldTypeValue {
		case util.TypeStructField:
			return fmt.Errorf("illegal primary key field type, type:%s", s.fieldTypeName)
		default:
		}
	}

	return nil
}

// Dump Dump
func (s *FieldInfo) Dump() string {
	return fmt.Sprintf("index:%d,name:%s,typeValue:%d, typeName:%s,tag:%s, pkgPath:%s", s.fieldIndex, s.fieldName, s.fieldTypeValue, s.fieldTypeName, s.fieldTag, s.fieldPkgPath)
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

// Verify Verify
func (s *Fields) Verify() error {
	if len(*s) == 0 {
		return fmt.Errorf("no defined Fields")
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
	for k, v := range *s {
		fmt.Printf("\tkey:%d, val:[%s]\n", k, v.Dump())
	}
}

// GetFieldInfo GetFieldInfo
func GetFieldInfo(idx int, fieldType *reflect.StructField, fieldVal *reflect.Value) *FieldInfo {
	info := &FieldInfo{}
	info.fieldIndex = idx
	info.fieldName = fieldType.Name
	info.fieldTag = newFieldTag(fieldType.Tag.Get("orm"))

	val := reflect.Indirect(*fieldVal)
	// 这里用val.Type()而不用fieldType.Type来判断是因为Field会是对象的指针，所以通过类型是判断不出真实类型的
	tVal, err := GetFieldType(val.Type())
	if err != nil {
		log.Printf("GetFieldType failed, idx:%d, name:%s, type:%s, err:%s", idx, fieldType.Name, fieldType.Type.Kind(), err.Error())
		return nil
	}

	info.fieldTypeValue = tVal
	info.fieldTypeName = val.Type().String()
	info.fieldValue = val
	info.fieldPkgPath = val.Type().PkgPath()

	return info
}
