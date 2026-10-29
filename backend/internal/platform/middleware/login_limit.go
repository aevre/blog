package middleware

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

	"blog-server/internal/platform/cache"
	"blog-server/internal/platform/logging"
	"blog-server/internal/shared/errx"
	"blog-server/internal/shared/httpx"

	"github.com/labstack/echo/v5"
)

type LoginLimiter interface {
	Allow(context.Context, string, int64, time.Duration) (bool, error)
}

// LoginLimit limits attempts per directly connected IP; Redis failures fail open.
func LoginLimit(store LoginLimiter, log logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if c.Request().Method != http.MethodPost || c.Request().URL.Path != "/api/v1/auth/login" {
				return next(c)
			}
			key := fmt.Sprintf("blog:login:%x", sha256.Sum256([]byte(c.RealIP())))
			allowed, err := store.Allow(c.Request().Context(), key, 10, time.Minute)
			if err != nil {
				log.Error("login limiter unavailable", logger.Err(err))
				return next(c)
			}
			if !allowed {
				return c.JSON(http.StatusTooManyRequests, httpx.Error(errx.CodeInvalidParam, "too many login attempts"))
			}
			return next(c)
		}
	}
}

var _ LoginLimiter = (*cache.Client)(nil)
