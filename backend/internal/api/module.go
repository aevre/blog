// Package api composes HTTP routes, health checks, and error handling.
package api

import (
	accounthandler "blog-server/internal/account/handler"
	setuphandler "blog-server/internal/bootstrap/handler"
	contenthandler "blog-server/internal/content/handler"
	linkhandler "blog-server/internal/link/handler"
	"blog-server/internal/platform/middleware"
	summaryhandler "blog-server/internal/summary/handler"

	"github.com/labstack/echo/v5"
	"go.uber.org/fx"
)

type Handlers struct {
	fx.In

	Auth      accounthandler.AuthHandler
	User      accounthandler.UserHandler
	Post      contenthandler.PostHandler
	Rss       contenthandler.RssHandler
	Link      linkhandler.LinkHandler
	Bootstrap setuphandler.BootstrapHandler
	Model     summaryhandler.ModelHandler
}

func RegisterRoutes(app *echo.Echo, handlers Handlers, auth *middleware.AuthMiddleware) {
	v1 := app.Group("/api").Group("/v1")
	setuphandler.RegisterBootstrapRouter(v1, handlers.Bootstrap)
	accounthandler.RegisterAuthRoutes(v1, handlers.Auth)
	contenthandler.RegisterPostRoutes(v1, handlers.Post, auth)
	contenthandler.RegisterRssRoutes(v1, handlers.Rss)
	linkhandler.RegisterLinkRoutes(v1, handlers.Link, auth)
	summaryhandler.RegisterModelRoutes(v1, handlers.Model)
	accounthandler.RegisterUserRoutes(v1, handlers.User, auth)
}

func Module() fx.Option {
	return fx.Module("api", fx.Invoke(RegisterRoutes, RegisterHealth))
}
