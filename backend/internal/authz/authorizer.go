package authz

import (
	"errors"

	accountmodel "blog-server/internal/account/model"
	"blog-server/internal/shared/errx"
)

type Authorizer interface {
	Require(role accountmodel.UserRole, resource Resource, action Action) error
}

type authorizer struct{}

func NewAuthorizer() Authorizer {
	return &authorizer{}
}

func ErrForbidden() error {
	return errx.New(errx.CodeForbidden, errors.New("permission denied"))
}

func (a *authorizer) Require(role accountmodel.UserRole, resource Resource, action Action) error {
	if !can(role, resource, action) {
		return ErrForbidden()
	}

	return nil
}
