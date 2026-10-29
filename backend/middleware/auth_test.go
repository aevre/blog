package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"blog-server/config"
	"blog-server/entity"
	"blog-server/pkg/errx"
	"blog-server/pkg/jwt"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

func TestAuthMiddlewareRejectsRefreshToken(t *testing.T) {
	cfg := &config.Config{JWT: config.JWTConfig{
		Secret:            "test-secret",
		Issuer:            "test",
		AccessExpiration:  time.Minute,
		RefreshExpiration: time.Hour,
	}}
	_, refresh, err := jwt.New(cfg.JWT).GenerateAllTokens(1, entity.UserRoleReader, uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	ctx := e.NewContext(httptest.NewRequest(http.MethodGet, "/private", nil), httptest.NewRecorder())
	ctx.Request().Header.Set("Authorization", "Bearer "+refresh)
	err = NewAuthMiddleware(cfg).Handler()(func(*echo.Context) error {
		t.Fatal("refresh token reached protected handler")
		return nil
	})(ctx)
	if got := errx.ToAppError(err).Code; got != errx.CodeUnauthorized {
		t.Fatalf("error code = %d, want %d", got, errx.CodeUnauthorized)
	}
}
