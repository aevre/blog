package logger

import (
	"context"

	"go.uber.org/fx"
)

// Module registers the logger implementation into the Fx dependency injection container.
//
// It provides a single Logger instance via NewLogger for application-wide use.
func Module() fx.Option {
	return fx.Module("logger",
		fx.Provide(NewLogger),
		fx.Invoke(registerLifecycle),
	)
}

func registerLifecycle(lc fx.Lifecycle, log Logger) {
	lc.Append(
		fx.Hook{
			OnStop: func(context.Context) error {
				return log.Sync()
			},
		},
	)
}
