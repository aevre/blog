package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"blog-server/internal/platform/config"
	"blog-server/internal/platform/logging"
	"github.com/labstack/echo/v5"
)

type limiterStub struct {
	allowed bool
	err     error
	calls   int
}

func (s *limiterStub) Allow(context.Context, string, int64, time.Duration) (bool, error) {
	s.calls++
	return s.allowed, s.err
}

func TestLoginLimit(t *testing.T) {
	for _, tc := range []struct {
		name    string
		allowed bool
		err     error
		path    string
		want    int
		calls   int
	}{
		{"allowed", true, nil, "/api/v1/auth/login", 204, 1},
		{"blocked", false, nil, "/api/v1/auth/login", 429, 1},
		{"redis failure", false, errors.New("unavailable"), "/api/v1/auth/login", 204, 1},
		{"other route", false, nil, "/api/v1/auth/refresh", 204, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := &limiterStub{allowed: tc.allowed, err: tc.err}
			e := echo.New()
			rec := httptest.NewRecorder()
			ctx := e.NewContext(httptest.NewRequest(http.MethodPost, tc.path, nil), rec)
			log := logger.NewLogger(&config.Config{Log: config.LogConfig{Level: "error", Format: "json"}})
			err := LoginLimit(store, log)(func(c *echo.Context) error { return c.NoContent(204) })(ctx)
			if err != nil || rec.Code != tc.want || store.calls != tc.calls {
				t.Fatalf("status=%d calls=%d err=%v", rec.Code, store.calls, err)
			}
		})
	}
}
