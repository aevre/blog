// Package account
package account

import (
	"blog-server/internal/account/handler"
	"blog-server/internal/account/repository"
	"blog-server/internal/account/service"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"account",
		fx.Provide(
			repository.NewUserRepo,
			repository.NewSessions,
			service.NewEmailService,
			service.NewAuthService,
			service.NewUserService,
			handler.NewAuthHandler,
			handler.NewUserHandler,
		),
	)
}
