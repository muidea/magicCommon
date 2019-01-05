package model

import (
	"fmt"
	"reflect"

	"muidea.com/magicCommon/orm/util"
)

type typeSlice struct {
	typeValue     int
	typeName      string
	typePkgPath   string
	typeIsPtr     bool
	typeDepend    reflect.Type
	typeDependPtr bool
}

func (s *typeSlice) Name() string {
	return s.typeName
}

func (s *typeSlice) Value() int {
	return s.typeValue
}

func (s *typeSlice) IsPtr() bool {
	return s.typeIsPtr
}

func (s *typeSlice) PkgPath() string {
	return s.typePkgPath
}

func (s *typeSlice) String() string {
	ret := fmt.Sprintf("val:%d,name:%s,pkgPath:%s,isPtr:%v", s.typeValue, s.typeName, s.typePkgPath, s.typeIsPtr)
	if s.typeDepend != nil {
		ret = fmt.Sprintf("%s,depend:[%s], dependPtr:[%v]", ret, s.typeDepend, s.typeDependPtr)
	}

	return ret
}

func (s *typeSlice) Depend() (reflect.Type, bool) {
	return s.typeDepend, s.typeDependPtr
}

func (s *typeSlice) Copy() FieldType {
	return &typeSlice{
		typeIsPtr:     s.typeIsPtr,
		typeName:      s.typeName,
		typePkgPath:   s.typePkgPath,
		typeValue:     s.typeValue,
		typeDepend:    s.typeDepend,
		typeDependPtr: s.typeDependPtr,
	}
}

func getSliceType(val reflect.Type, isPtr bool) (ret FieldType, err error) {
	tVal, tErr := util.GetTypeValueEnum(val)
	if tErr != nil {
		err = tErr
		return
	}

	var typeDepend reflect.Type
	typeDependPtr := false
	if util.IsSliceType(tVal) {
		sliceVal := val.Elem()
		sliceRawVal := sliceVal
		if sliceRawVal.Kind() == reflect.Ptr {
			sliceRawVal = sliceRawVal.Elem()
			typeDependPtr = true
		}

		tSliceVal, tSliceErr := util.GetTypeValueEnum(sliceRawVal)
		if tSliceErr != nil {
			err = tSliceErr
			return
		}
		if util.IsSliceType(tSliceVal) {
			err = fmt.Errorf("illegal slice depend type, type:[%s]", sliceRawVal.String())
			return
		}

		if util.IsStructType(tSliceVal) {
			typeDepend = sliceRawVal
		}

		ret = &typeSlice{typeValue: tVal, typeName: val.String(), typePkgPath: val.PkgPath(), typeIsPtr: isPtr, typeDepend: typeDepend, typeDependPtr: typeDependPtr}
		return
	}

	err = fmt.Errorf("illegal slice type, type:%s", val.String())

	return
}
