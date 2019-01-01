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

func newFieldType(sf reflect.StructField) (ret FieldType, err error) {
	val := sf.Type
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	tVal, tErr := util.GetTypeValueEnum(val)
	if tErr != nil {
		err = tErr
		return
	}
	if util.IsBasicType(tVal) {
		ret, err = getBasicType(val)
		return
	}

	if util.IsStructType(tVal) {
		ret, err = getStructType(val)
		return
	}

	if util.IsSliceType(tVal) {
		ret, err = getSliceType(val)
		return
	}

	err = fmt.Errorf("illegal fieldType, type:%s", val.String())
	return
}
