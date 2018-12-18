package model

import (
	"fmt"
	"reflect"
)

// FieldType FieldType
type FieldType interface {
	Name() string
	Value() int
	PkgPath() string
	String() string
}

type typeImpl struct {
	typeValue   int
	typeName    string
	typePkgPath string
}

func (s *typeImpl) Name() string {
	return s.typeName
}

func (s *typeImpl) Value() int {
	return s.typeValue
}

func (s *typeImpl) PkgPath() string {
	return s.typePkgPath
}

func (s *typeImpl) String() string {
	return fmt.Sprintf("val:%d,name:%s,pkgPath:%s", s.typeValue, s.typeName, s.typePkgPath)
}

func newFieldType(val reflect.Value) FieldType {
	tVal, err := GetFieldType(val.Type())
	if err != nil {
		msg := fmt.Sprintf("get field type failed, err:%s", err.Error())
		panic(msg)
	}

	return &typeImpl{typeValue: tVal, typeName: val.Type().String(), typePkgPath: val.Type().PkgPath()}
}
