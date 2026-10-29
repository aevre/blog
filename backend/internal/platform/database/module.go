// Package database provides PostgreSQL connections, Ent clients, and transaction support.
package database

import (
	"context"
	"time"

	"blog-server/internal/platform/config"
	"blog-server/internal/shared/txmgr"

	"go.uber.org/fx"
)

// Module returns an fx.Module that registers database dependencies,
// including the database client, transaction manager, and lifecycle hooks.
func Module() fx.Option {
	return fx.Module("database",
		fx.Provide(
			NewDataStore,
			func(ds *DataStore) txmgr.TxManager { return ds },
		),
		fx.Invoke(registerLifecycle),
	)
}

// registerLifecycle attaches lifecycle hooks for the database module.
//
// It ensures that the database connection is properly closed when the
// application stops.
func registerLifecycle(lc fx.Lifecycle, ds *DataStore, cfg *config.Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			timeout := cfg.Database.Timeout
			if timeout <= 0 {
				timeout = 5 * time.Second
			}
			pingCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			return ds.DB().PingContext(pingCtx)
		},
		OnStop: func(context.Context) error {
			return ds.Close()
		},
	})
}
