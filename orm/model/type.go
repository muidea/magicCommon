package model

import (
	"fmt"
	"reflect"
)

// FieldType FieldType
type FieldType interface {
	Name() string
	Value() int
	IsReference() bool
	PkgPath() string
	String() string
}

type typeImpl struct {
	typeValue   int
	typeName    string
	typePkgPath string
	isReference bool
}

func (s *typeImpl) Name() string {
	return s.typeName
}

func (s *typeImpl) Value() int {
	return s.typeValue
}

func (s *typeImpl) IsReference() bool {
	return s.isReference
}

func (s *typeImpl) PkgPath() string {
	return s.typePkgPath
}

func (s *typeImpl) String() string {
	return fmt.Sprintf("val:%d,name:%s,pkgPath:%s, isReference:%v", s.typeValue, s.typeName, s.typePkgPath, s.isReference)
}

func newFieldType(sf *reflect.StructField) FieldType {
	val := sf.Type

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	tVal, err := GetFieldType(val)
	if err != nil {
		msg := fmt.Sprintf("get field type failed, err:%s", err.Error())
		panic(msg)
	}

	isReference := IsReference(val)

	return &typeImpl{typeValue: tVal, typeName: val.String(), typePkgPath: val.PkgPath(), isReference: isReference}
}
