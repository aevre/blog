// Package setup wires initial site setup into the application.
package setup

import (
	"blog-server/internal/setup/handler"
	"blog-server/internal/setup/repository"
	"blog-server/internal/setup/service"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"setup",
		fx.Provide(
			repository.NewSiteRepo,
			repository.NewSystemRepo,
			service.NewSetupService,
			handler.NewSetupHandler,
		),
	)
}
