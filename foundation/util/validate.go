package util

import "reflect"

// ValidateFunc 校验是否是函数
func ValidateFunc(fun interface{}) {
	if reflect.TypeOf(fun).Kind() != reflect.Func {
		panic("fun must be a callable func")
	}
}

// ValidatePtr 校验是否是指针
func ValidatePtr(ptr interface{}) {
	if reflect.TypeOf(ptr).Kind() != reflect.Ptr {
		panic("fun must be a object ptr")
	}
}
