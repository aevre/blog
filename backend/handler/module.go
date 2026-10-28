// Package handler
package handler

import (
	"blog-server/middleware"

	"github.com/labstack/echo/v5"
	"go.uber.org/fx"
)

type Handlers struct {
	fx.In

	Post  PostHandler
	Rss   RssHandler
	Auth  AuthHandler
	Link  LinkHandler
	Model ModelHandler
	Setup SetupHandler
	User  UserHandler
}

type Middlewares struct {
	fx.In

	Auth *middleware.AuthMiddleware
}

func RegisterRoutes(
	app *echo.Echo,
	h Handlers,
	m Middlewares,
) {
	api := app.Group("/api")
	v1 := api.Group("/v1")
	RegisterSetupRouter(v1, h.Setup)
	RegisterAuthRoutes(v1, h.Auth)
	RegisterPostRoutes(v1, h.Post, m.Auth)
	RegisterRssRoutes(v1, h.Rss)
	RegisterLinkRoutes(v1, h.Link)
	RegisterModelRoutes(v1, h.Model)
	RegisterUserRoutes(v1, h.User, m.Auth)
}

func Module() fx.Option {
	return fx.Module(
		"handler",
		fx.Provide(
			NewSetupHandler,
			NewPostHandler,
			NewRssHandler,
			NewAuthHandler,
			NewLinkHandler,
			NewModelHandler,
			NewUserHandler,
		),
		fx.Invoke(
			RegisterRoutes,
		),
	)
}
