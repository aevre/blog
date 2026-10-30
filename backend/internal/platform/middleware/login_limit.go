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

var loginLimits = []struct {
	name   string
	limit  int64
	window time.Duration
}{
	{name: "minute", limit: 10, window: time.Minute},
	{name: "hour", limit: 100, window: time.Hour},
}

// LoginLimit limits attempts per directly connected IP; Redis failures fail open.
func LoginLimit(store LoginLimiter, log logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if c.Request().Method != http.MethodPost || c.Request().URL.Path != "/api/v1/auth/login" {
				return next(c)
			}

			ipHash := sha256.Sum256([]byte(c.RealIP()))
			for _, limit := range loginLimits {
				key := fmt.Sprintf("blog:login:%s:%x", limit.name, ipHash)
				allowed, err := store.Allow(c.Request().Context(), key, limit.limit, limit.window)
				if err != nil {
					log.Error("login limiter unavailable", logger.Err(err))
					return next(c)
				}
				if !allowed {
					return c.JSON(http.StatusTooManyRequests, httpx.Error(errx.CodeInvalidParam, "too many login attempts"))
				}
			}

			return next(c)
		}
	}
}

var _ LoginLimiter = (*cache.Client)(nil)
