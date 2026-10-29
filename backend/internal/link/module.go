// Package link wires the link feature into the application.
package link

import (
	"blog-server/internal/link/handler"
	"blog-server/internal/link/repository"
	"blog-server/internal/link/service"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"link",
		fx.Provide(
			repository.NewLinkRepo,
			service.NewLinkService,
			handler.NewLinkHandler,
		),
	)
}
