// Package bootstrap wires initial site setup into the application.
package bootstrap

import (
	"blog-server/internal/bootstrap/handler"
	"blog-server/internal/bootstrap/repository"
	"blog-server/internal/bootstrap/service"

	"go.uber.org/fx"
)

func registerBootstrapLifecycle(lc fx.Lifecycle, svc service.BootstrapService) {
	lc.Append(fx.Hook{
		OnStart: svc.EnsureBootstrapToken,
	})
}

func Module() fx.Option {
	return fx.Module(
		"bootstrap",
		fx.Provide(
			repository.NewSiteRepo,
			repository.NewSystemRepo,
			service.NewBootstrapService,
			handler.NewBootstrapHandler,
		),
		fx.Invoke(registerBootstrapLifecycle),
	)
}
