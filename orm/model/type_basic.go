package model

import (
	"fmt"
	"reflect"

	"muidea.com/magicCommon/orm/util"
)

type typeBasic struct {
	typeValue   int
	typeName    string
	typePkgPath string
	typeIsPtr   bool
}

func (s *typeBasic) Name() string {
	return s.typeName
}

func (s *typeBasic) Value() int {
	return s.typeValue
}

func (s *typeBasic) IsPtr() bool {
	return s.typeIsPtr
}

func (s *typeBasic) PkgPath() string {
	return s.typePkgPath
}

func (s *typeBasic) String() string {
	ret := fmt.Sprintf("val:%d,name:%s,pkgPath:%s,isPtr:%v", s.typeValue, s.typeName, s.typePkgPath, s.typeIsPtr)

	return ret
}

func (s *typeBasic) Depend() reflect.Type {
	return nil
}

func getBasicType(val reflect.Type) (ret FieldType, err error) {
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
	if util.IsBasicType(tVal) {
		ret = &typeBasic{typeValue: tVal, typeName: val.String(), typePkgPath: val.PkgPath(), typeIsPtr: isPtr}
		return
	}

	err = fmt.Errorf("illegal basic type, type:%s", val.String())

	return
}
