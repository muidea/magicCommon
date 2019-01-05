package model

import (
	"fmt"
	"reflect"

	"muidea.com/magicCommon/orm/util"
)

type typeStruct struct {
	typeValue     int
	typeName      string
	typePkgPath   string
	typeIsPtr     bool
	typeDepend    reflect.Type
	typeDependPtr bool
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
		ret = fmt.Sprintf("%s,depend:[%s],typeDependPtr:[%v]", ret, s.typeDepend, s.typeDependPtr)
	}

	return ret
}

func (s *typeStruct) Depend() (reflect.Type, bool) {
	return s.typeDepend, s.typeDependPtr
}

func (s *typeStruct) Copy() FieldType {
	return &typeSlice{
		typeIsPtr:     s.typeIsPtr,
		typeName:      s.typeName,
		typePkgPath:   s.typePkgPath,
		typeValue:     s.typeValue,
		typeDepend:    s.typeDepend,
		typeDependPtr: s.typeDependPtr,
	}
}

func getStructType(val reflect.Type, isPtr bool) (ret FieldType, err error) {
	tVal, tErr := util.GetTypeValueEnum(val)
	if tErr != nil {
		err = tErr
		return
	}

	if util.IsStructType(tVal) {
		ret = &typeStruct{typeValue: tVal, typeName: val.String(), typePkgPath: val.PkgPath(), typeIsPtr: isPtr, typeDepend: val, typeDependPtr: isPtr}
		return
	}

	err = fmt.Errorf("illegal struct type, type:%s", val.String())

	return
}
