package model

import (
	"fmt"
	"reflect"

	"muidea.com/magicCommon/orm/util"
)

type typeStruct struct {
	typeValue   int
	typeName    string
	typePkgPath string
	typeIsPtr   bool
	typeDepend  reflect.Type
}

func (s *typeStruct) Name() string {
	return s.typeName
}

func (s *typeStruct) Value() int {
	return s.typeValue
}

func (s *typeStruct) IsPtr() bool {
	return s.typeIsPtr
}

func (s *typeStruct) PkgPath() string {
	return s.typePkgPath
}

func (s *typeStruct) String() string {
	ret := fmt.Sprintf("val:%d,name:%s,pkgPath:%s,isPtr:%v", s.typeValue, s.typeName, s.typePkgPath, s.typeIsPtr)
	if s.typeDepend != nil {
		ret = fmt.Sprintf("%s,depend:[%s]", ret, s.typeDepend)
	}

	return ret
}

func (s *typeStruct) Depend() reflect.Type {
	return s.typeDepend
}

func getStructType(val reflect.Type) (ret FieldType, err error) {
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
	if util.IsBasicType(tVal) {
		ret = &typeStruct{typeValue: tVal, typeName: val.String(), typePkgPath: val.PkgPath(), typeIsPtr: isPtr}
		return
	}

	var typeDepend reflect.Type
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
			typeDepend = sliceVal
		}

		ret = &typeStruct{typeValue: tVal, typeName: val.String(), typePkgPath: val.PkgPath(), typeIsPtr: isPtr, typeDepend: typeDepend}
		return
	}

	if util.IsStructType(tVal) {
		ret = &typeStruct{typeValue: tVal, typeName: val.String(), typePkgPath: val.PkgPath(), typeIsPtr: isPtr, typeDepend: rawVal}
		return
	}

	return
}
