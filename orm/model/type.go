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
	Depend() reflect.Type
	Copy() FieldType
}

func newFieldType(sf reflect.StructField) (ret FieldType, err error) {
	val := sf.Type

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
		ret, err = getBasicType(rawVal, isPtr)
		return
	}

	if util.IsStructType(tVal) {
		ret, err = getStructType(rawVal, isPtr)
		return
	}

	if util.IsSliceType(tVal) {
		ret, err = getSliceType(rawVal, isPtr)
		return
	}

	err = fmt.Errorf("illegal fieldType, type:%s", val.String())
	return
}
