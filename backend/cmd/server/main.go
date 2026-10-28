package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"blog-server/authz"
	"blog-server/cache"
	"blog-server/config"
	"blog-server/datastore"
	"blog-server/handler"
	"blog-server/logger"
	"blog-server/middleware"
	"blog-server/migration"
	"blog-server/pkg/validatorx"
	"blog-server/repository"
	"blog-server/scheduler"
	"blog-server/service"

	"github.com/labstack/echo/v5"
	"go.uber.org/fx"
)

// main initializes the application using Uber Fx and starts the dependency
// injection container.
//
// It registers configuration, logging, datastore, migration, HTTP server,
// and other application modules required to run the service.
func main() {
	app := fx.New(
		fx.Options(
			config.Module(),
			logger.Module(),
			cache.Module(),
			datastore.Module(),
			migration.Module(),
			repository.Module(),
			authz.Module(),
			service.Module(),
			handler.Module(),
			scheduler.Module(),
		),
		fx.Provide(
			validatorx.NewValidator,
			middleware.NewAuthMiddleware,
			providerEchoApp,
		),
		fx.Invoke(
			runServerLifecycle,
		),
	)

	app.Run()
}

func providerEchoApp(cfg *config.Config, log logger.Logger) *echo.Echo {
	echoCfg := echo.Config{
		HTTPErrorHandler: handler.ErrorHandler(cfg, log),
		IPExtractor:      echo.ExtractIPDirect(),
	}

	app := echo.NewWithConfig(echoCfg)
	app.Use(middleware.RequestLogger(log))
	app.Use(middleware.Recover)
	app.Use(middleware.BodyLimit(10 * 1024 * 1024))

	return app
}

// runServerLifecycle registers Echo server startup and graceful shutdown
// hooks into the Fx application lifecycle.
//
// OnStart:
//   - Starts the HTTP server in a separate goroutine
//
// OnStop:
//   - Performs graceful shutdown using configured timeout
//   - Ensures in-flight requests are completed before exit
func runServerLifecycle(
	lc fx.Lifecycle,
	app *echo.Echo,
	cfg *config.Config,
	log logger.Logger,
) {
	srv := &http.Server{
		Addr:              cfg.Server.Addr(),
		Handler:           app,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		MaxHeaderBytes:    cfg.Server.MaxHeaderBytes,
		ReadHeaderTimeout: cfg.Server.ReadTimeout,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			listener, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			log.Info("server started", logger.String("address", srv.Addr))
			go func() {
				if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Error("server failed", logger.Err(err))
				}
			}()
			return nil
		},

		OnStop: func(ctx context.Context) error {
			timeout := cfg.Server.GracefulShutdown
			if timeout <= 0 {
				timeout = 5 * time.Second
			}
			shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			log.Info("server stopping")
			if err := srv.Shutdown(shutdownCtx); err != nil {
				return err
			}
			log.Info("server stopped")
			return nil
		},
	})
}
