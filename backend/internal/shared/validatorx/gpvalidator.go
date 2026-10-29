package validatorx

import "github.com/go-playground/validator/v10"

type gpValidator struct {
	v *validator.Validate
}

func NewValidator() Validator {
	v := validator.New(validator.WithRequiredStructEnabled())
	return &gpValidator{v: v}
}

func (g *gpValidator) RegisterTagNameFunc(fn validator.TagNameFunc) {
	g.v.RegisterTagNameFunc(fn)
}

func (g *gpValidator) RegisterValidation(tag string, fn validator.Func) error {
	return g.v.RegisterValidation(tag, fn)
}

func (g *gpValidator) Struct(s any) error {
	return g.v.Struct(s)
}

func (g *gpValidator) Var(field any, tag string) error {
	return g.v.Var(field, tag)
}
