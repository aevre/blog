package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"blog-server/config"
	"blog-server/logger"
	"blog-server/middleware"
	"blog-server/pkg/errx"

	"github.com/labstack/echo/v5"
)

type errorLogRecorder struct {
	logger.Logger
	fields  []logger.Field
	entries [][]logger.Field
}

func (l *errorLogRecorder) With(fields ...logger.Field) logger.Logger {
	l.fields = append(l.fields, fields...)
	return l
}

func (l *errorLogRecorder) WithCallerSkip(int) logger.Logger { return l }

func (l *errorLogRecorder) Warn(_ string, fields ...logger.Field) {
	l.entries = append(l.entries, append(l.fields, fields...))
}

func (l *errorLogRecorder) Error(msg string, fields ...logger.Field) {
	l.Warn(msg, fields...)
}

func TestErrorHandlerLogDetailsByEnvironment(t *testing.T) {
	for _, env := range []string{config.EnvDev, config.EnvProd} {
		for _, status := range []int{http.StatusUnauthorized, http.StatusInternalServerError} {
			t.Run(env+"/"+http.StatusText(status), func(t *testing.T) {
				log := &errorLogRecorder{}
				cfg := &config.Config{App: config.AppConfig{Environment: env}}
				e := echo.New()
				e.HTTPErrorHandler = ErrorHandler(cfg, log)
				e.Use(middleware.RequestLogger(log))
				e.GET("/test", func(c *echo.Context) error {
					code := errx.CodeUnauthorized
					if status == http.StatusInternalServerError {
						code = errx.CodeInternalError
					}
					return errx.New(code, errors.New("diagnostic reason"))
				})
				res := httptest.NewRecorder()
				e.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/test", nil))
				if res.Code != status || len(log.entries) != 1 {
					t.Fatalf("status = %d, log entries = %d", res.Code, len(log.entries))
				}
				fields := map[string]any{}
				for _, field := range log.entries[0] {
					if _, exists := fields[field.Key]; exists {
						t.Fatalf("duplicate field %q", field.Key)
					}
					fields[field.Key] = field.Value
				}
				_, hasStack := fields["stack"]
				_, hasError := fields["error"]
				if hasStack != (env == config.EnvDev || status >= 500) || !hasError {
					t.Fatalf("unexpected diagnostic fields: %#v", fields)
				}
				wantCode := errx.CodeUnauthorized
				if status == http.StatusInternalServerError {
					wantCode = errx.CodeInternalError
				}
				if fields["code"] != wantCode || !strings.Contains(fields["error"].(error).Error(), "diagnostic reason") {
					t.Fatalf("missing error details: %#v", fields)
				}
				if hasStack && !strings.Contains(fields["stack"].(string), "TestErrorHandlerLogDetailsByEnvironment") {
					t.Fatal("stack does not identify the error origin")
				}
				if strings.Contains(res.Body.String(), "diagnostic reason") {
					t.Fatal("internal error leaked into HTTP response")
				}
			})
		}
	}
}
