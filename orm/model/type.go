package model

import (
	"fmt"
	"reflect"

	"muidea.com/magicCommon/orm/util"
)

// FieldType FieldType
type FieldType interface {
	Name() string
	Value() int
	IsPtr() bool
	PkgPath() string
	String() string
	Depend() FieldType
}

type typeImpl struct {
	typeValue   int
	typeName    string
	typePkgPath string
	typeIsPtr   bool
	typeDepend  FieldType
}

func (s *typeImpl) Name() string {
	return s.typeName
}

func (s *typeImpl) Value() int {
	return s.typeValue
}

func (s *typeImpl) IsPtr() bool {
	return s.typeIsPtr
}

func (s *typeImpl) PkgPath() string {
	return s.typePkgPath
}

func (s *typeImpl) String() string {
	ret := fmt.Sprintf("val:%d,name:%s,pkgPath:%s,isPtr:%v", s.typeValue, s.typeName, s.typePkgPath, s.typeIsPtr)
	if s.typeDepend != nil {
		ret = fmt.Sprintf("%s,depend:[%s]", ret, s.typeDepend)
	}

	return ret
}

func (s *typeImpl) Depend() FieldType {
	return s.typeDepend
}

func getFieldType(val reflect.Type) (ret FieldType, err error) {
	isPtr := false
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		isPtr = true
	}

	tVal, tErr := util.GetTypeValueEnum(val)
	if tErr != nil {
		err = tErr
		return
	}

	var typeDepend FieldType
	if util.IsSliceType(tVal) {
		isSlicePtr := false
		sliceVal := val.Elem()
		if sliceVal.Kind() == reflect.Ptr {
			sliceVal = sliceVal.Elem()
			isSlicePtr = true
		}

		tSliceVal, tSliceErr := util.GetTypeValueEnum(sliceVal)
		if tSliceErr != nil {
			err = tSliceErr
			return
		}
		if util.IsStructType(tSliceVal) {
			typeDepend = &typeImpl{typeValue: tSliceVal, typeName: sliceVal.String(), typePkgPath: sliceVal.PkgPath(), typeIsPtr: isSlicePtr}
		}
		if util.IsSliceType(tSliceVal) {
			err = fmt.Errorf("illegal slice depend type")
			return
		}
	}

	ret = &typeImpl{typeValue: tVal, typeName: val.String(), typePkgPath: val.PkgPath(), typeIsPtr: isPtr, typeDepend: typeDepend}
	return
}

func newFieldType(sf reflect.StructField) (ret FieldType, err error) {
	val := sf.Type
	ret, err = getFieldType(val)

	return
}
