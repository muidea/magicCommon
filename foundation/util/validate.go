package util

import (
	sysValidator "github.com/go-playground/validator/v10"
	"reflect"
)

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

type Validator interface {
	Validate(value interface{}) error
}

func NewFormValidator() Validator {
	return &validator{validate: sysValidator.New()}
}

type validator struct {
	validate *sysValidator.Validate
}

func (s *validator) Validate(value interface{}) error {
	return s.validate.Struct(value)
}
