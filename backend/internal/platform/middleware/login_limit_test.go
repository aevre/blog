package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"blog-server/internal/platform/config"
	"blog-server/internal/platform/logging"
	"github.com/labstack/echo/v5"
)

type limiterResult struct {
	allowed bool
	err     error
}

type limiterCall struct {
	key    string
	limit  int64
	window time.Duration
}

type limiterStub struct {
	results []limiterResult
	calls   []limiterCall
}

func (s *limiterStub) Allow(_ context.Context, key string, limit int64, window time.Duration) (bool, error) {
	s.calls = append(s.calls, limiterCall{key: key, limit: limit, window: window})
	result := s.results[len(s.calls)-1]
	return result.allowed, result.err
}

func TestLoginLimit(t *testing.T) {
	unavailable := errors.New("unavailable")
	for _, tc := range []struct {
		name       string
		results    []limiterResult
		path       string
		wantStatus int
		wantCalls  int
	}{
		{"allowed", []limiterResult{{allowed: true}, {allowed: true}}, "/api/v1/auth/login", 204, 2},
		{"minute blocked", []limiterResult{{allowed: false}}, "/api/v1/auth/login", 429, 1},
		{"hour blocked", []limiterResult{{allowed: true}, {allowed: false}}, "/api/v1/auth/login", 429, 2},
		{"minute redis failure", []limiterResult{{err: unavailable}}, "/api/v1/auth/login", 204, 1},
		{"hour redis failure", []limiterResult{{allowed: true}, {err: unavailable}}, "/api/v1/auth/login", 204, 2},
		{"other route", nil, "/api/v1/auth/refresh", 204, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := &limiterStub{results: tc.results}
			e := echo.New()
			rec := httptest.NewRecorder()
			ctx := e.NewContext(httptest.NewRequest(http.MethodPost, tc.path, nil), rec)
			log := logger.NewLogger(&config.Config{Log: config.LogConfig{Level: "error", Format: "json"}})
			err := LoginLimit(store, log)(func(c *echo.Context) error { return c.NoContent(204) })(ctx)
			if err != nil || rec.Code != tc.wantStatus || len(store.calls) != tc.wantCalls {
				t.Fatalf("status=%d calls=%d err=%v", rec.Code, len(store.calls), err)
			}

			for i, call := range store.calls {
				limit := loginLimits[i]
				if call.limit != limit.limit || call.window != limit.window {
					t.Fatalf("call %d limit=%d window=%s", i, call.limit, call.window)
				}
				if !strings.HasPrefix(call.key, "blog:login:"+limit.name+":") {
					t.Fatalf("call %d key = %q", i, call.key)
				}
			}
		})
	}
}
