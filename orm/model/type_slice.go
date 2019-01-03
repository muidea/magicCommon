package model

import (
	"fmt"
	"reflect"

	"muidea.com/magicCommon/orm/util"
)

type typeSlice struct {
	typeValue   int
	typeName    string
	typePkgPath string
	typeIsPtr   bool
	typeDepend  reflect.Type
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
		ret = fmt.Sprintf("%s,depend:[%s]", ret, s.typeDepend)
	}

	return ret
}

func (s *typeSlice) Depend() reflect.Type {
	return s.typeDepend
}

func getSliceType(val reflect.Type, isPtr bool) (ret FieldType, err error) {
	tVal, tErr := util.GetTypeValueEnum(val)
	if tErr != nil {
		err = tErr
		return
	}

	var typeDepend reflect.Type
	if util.IsSliceType(tVal) {
		sliceVal := val.Elem()
		sliceRawVal := sliceVal
		if sliceRawVal.Kind() == reflect.Ptr {
			sliceRawVal = sliceRawVal.Elem()
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

		ret = &typeSlice{typeValue: tVal, typeName: val.String(), typePkgPath: val.PkgPath(), typeIsPtr: isPtr, typeDepend: typeDepend}
		return
	}

	err = fmt.Errorf("illegal slice type, type:%s", val.String())

	return
}
