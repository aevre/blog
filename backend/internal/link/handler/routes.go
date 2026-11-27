package handler

import (
	"blog-server/internal/platform/middleware"

	"github.com/labstack/echo/v5"
)

// RegisterLinkRoutes registers all link-related routes.
func RegisterLinkRoutes(r *echo.Group, h LinkHandler, auth *middleware.AuthMiddleware) {
	public := r.Group("/links")
	public.GET("", h.ListPublic)

	signed := r.Group("/links", auth.Handler())
	signed.POST("", h.Create)
	signed.GET("/mine/:id", h.GetMineByID)
	signed.PATCH("/mine/:id", h.Update)
	signed.DELETE("/mine/:id", h.Delete)

	admin := r.Group("/admin/links", auth.Handler())
	admin.GET("", h.AdminList)
	admin.GET("/:id", h.AdminGetByID)
	admin.PATCH("/:id", h.AdminUpdate)
	admin.DELETE("/:id", h.AdminDelete)
	admin.POST("/:id/review", h.Review)
}
