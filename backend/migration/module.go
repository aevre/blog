package migration

import (
	"context"

	"blog-server/internal/platform/database"

	"go.uber.org/fx"
)

// runMigrationLifecycle is optional: either register it with the root fx.Invoke
// before runServerLifecycle, or use fx.Module, since Invokes in child modules
// are executed before Invokes in their parent module.
func runMigrationLifecycle(
	lc fx.Lifecycle,
	ds *database.DataStore,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return Up(ctx, ds.DB())
		},
	})
}

func Module() fx.Option {
	return fx.Module(
		"migration",
		fx.Invoke(
			runMigrationLifecycle,
		),
	)
}
