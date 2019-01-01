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
	typeDepend  FieldType
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

func (s *typeSlice) Depend() FieldType {
	return s.typeDepend
}

func getSliceType(val reflect.Type) (ret FieldType, err error) {
	isPtr := false
	rawVal := val
	if rawVal.Kind() == reflect.Ptr {
		rawVal = rawVal.Elem()
		isPtr = true
	}

	tVal, tErr := util.GetTypeValueEnum(rawVal)
	if tErr != nil {
		err = tErr
		return
	}

	var typeDepend FieldType
	if util.IsSliceType(tVal) {
		sliceVal := val.Elem()
		if sliceVal.Kind() == reflect.Ptr {
			sliceVal = sliceVal.Elem()
		}

		tSliceVal, tSliceErr := util.GetTypeValueEnum(sliceVal)
		if tSliceErr != nil {
			err = tSliceErr
			return
		}
		if util.IsSliceType(tSliceVal) {
			err = fmt.Errorf("illegal slice depend type, type:[%s]", sliceVal.String())
			return
		}

		if util.IsStructType(tSliceVal) {
			typeDepend = &typeSlice{typeValue: tSliceVal, typeName: sliceVal.String(), typePkgPath: sliceVal.PkgPath()}
		}

		ret = &typeSlice{typeValue: tVal, typeName: val.String(), typePkgPath: val.PkgPath(), typeIsPtr: isPtr, typeDepend: typeDepend}
		return
	}

	err = fmt.Errorf("illegal slice type, type:%s", val.String())

	return
}
