// Package content wires the content feature into the application.
package content

import (
	"blog-server/internal/content/handler"
	"blog-server/internal/content/repository"
	"blog-server/internal/content/service"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"content",
		fx.Provide(
			repository.NewPostRepo,
			service.NewPostService,
			service.NewRssService,
			handler.NewPostHandler,
			handler.NewRssHandler,
		),
	)
}
