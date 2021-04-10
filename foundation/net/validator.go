package net

import sysValidator "github.com/go-playground/validator/v10"

type Validator interface {
	Validate(value interface{}) error
}

func NewValidator() Validator {
	return &validator{validate: sysValidator.New()}
}

type validator struct {
	validate *sysValidator.Validate
}

func (s *validator) Validate(value interface{}) error {
	return s.validate.Struct(value)
}
