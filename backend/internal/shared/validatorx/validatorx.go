// Package validatorx
package validatorx

import (
	"github.com/go-playground/validator/v10"
)

type Validator interface {
	Struct(s any) error
	Var(field any, tag string) error
	RegisterValidation(tag string, fn validator.Func) error
	RegisterTagNameFunc(fn validator.TagNameFunc)
}
