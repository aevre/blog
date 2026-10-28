package middleware

import (
	"fmt"
	"strings"

	"blog-server/config"
	"blog-server/contextx"
	"blog-server/pkg/errx"
	"blog-server/pkg/jwt"

	"github.com/labstack/echo/v5"
)

// AuthMiddleware handles JWT authentication for protected routes
type AuthMiddleware struct {
	jwt jwt.Jwt
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(cfg *config.Config) *AuthMiddleware {
	return &AuthMiddleware{
		jwt: jwt.New(cfg.JWT),
	}
}

// Handler returns a Echo handler that validates JWT tokens
// Optional roles can be specified for role-based access control
func (m *AuthMiddleware) Handler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return errx.New(errx.CodeUnauthorized, fmt.Errorf("missing authorization header"))
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return errx.New(errx.CodeUnauthorized, fmt.Errorf("invalid authorization header"))
			}

			tokenStr := parts[1]
			if tokenStr == "" {
				return errx.New(errx.CodeUnauthorized, fmt.Errorf("empty token"))
			}

			claims, err := m.jwt.Parse(tokenStr)
			if err != nil {
				return err
			}
			if claims.TokenType != jwt.TokenTypeAccess {
				return errx.New(errx.CodeUnauthorized, fmt.Errorf("access token required"))
			}

			user := contextx.User{
				ID:   claims.ID,
				Role: claims.Role,
			}

			ctx := contextx.SetUser(c.Request().Context(), user)

			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
