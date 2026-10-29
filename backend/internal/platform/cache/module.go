package cache

import (
	"context"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"cache",
		fx.Provide(
			NewCacheClient,
		),
		fx.Invoke(
			registerLifecycle,
		),
	)
}

func registerLifecycle(lc fx.Lifecycle, rc *Client) {
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return rc.Close()
		},
	})
}
