package api

import (
	"context"
	"net/http"
	"time"

	"blog-server/internal/platform/cache"
	"blog-server/internal/platform/database"

	"github.com/labstack/echo/v5"
)

func RegisterHealth(app *echo.Echo, ds *database.DataStore, rc *cache.Client) {
	app.GET("/health/live", func(c *echo.Context) error { return c.JSON(http.StatusOK, map[string]string{"status": "ok"}) })
	app.GET("/health/ready", func(c *echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()
		if err := ds.DB().PingContext(ctx); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
		}
		if err := rc.Ping(ctx); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
}
