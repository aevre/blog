// Package summary wires content summarization into the application.
package summary

import (
	"blog-server/internal/summary/handler"
	"blog-server/internal/summary/service"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"summary",
		fx.Provide(
			service.NewModelService,
			handler.NewModelHandler,
		),
	)
}
