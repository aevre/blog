package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"blog-server/contextx"
	"blog-server/logger"

	"github.com/labstack/echo/v5"
)

type logEntry struct {
	level  string
	msg    string
	fields []logger.Field
}

type testLogger struct {
	entries *[]logEntry
	fields  []logger.Field
}

func newTestLogger() *testLogger {
	entries := []logEntry{}
	return &testLogger{entries: &entries}
}

func (l *testLogger) record(level, msg string, fields ...logger.Field) {
	all := append(append([]logger.Field{}, l.fields...), fields...)
	*l.entries = append(*l.entries, logEntry{level: level, msg: msg, fields: all})
}
func (l *testLogger) Debug(msg string, fields ...logger.Field) { l.record("debug", msg, fields...) }
func (l *testLogger) Info(msg string, fields ...logger.Field)  { l.record("info", msg, fields...) }
func (l *testLogger) Warn(msg string, fields ...logger.Field)  { l.record("warn", msg, fields...) }
func (l *testLogger) Error(msg string, fields ...logger.Field) { l.record("error", msg, fields...) }
func (l *testLogger) Fatal(msg string, fields ...logger.Field) { l.record("fatal", msg, fields...) }
func (l *testLogger) Panic(msg string, fields ...logger.Field) { panic(msg) }
func (l *testLogger) With(fields ...logger.Field) logger.Logger {
	return &testLogger{entries: l.entries, fields: append(append([]logger.Field{}, l.fields...), fields...)}
}
func (l *testLogger) WithCallerSkip(int) logger.Logger          { return l }
func (l *testLogger) WithContext(context.Context) logger.Logger { return l }
func (l *testLogger) Sync() error                               { return nil }

func TestRequestLoggerCorrelatesRequestAndUsesRoutePattern(t *testing.T) {
	log := newTestLogger()
	e := echo.New()
	e.Use(RequestLogger(log))
	e.GET("/users/:id", func(c *echo.Context) error {
		if got := contextx.GetRequestID(c.Request().Context()); got != "request-1" {
			t.Fatalf("request ID in context = %q", got)
		}
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/users/42?token=secret", nil)
	req.Header.Set(echo.HeaderXRequestID, "request-1")
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)

	if got := res.Header().Get(echo.HeaderXRequestID); got != "request-1" {
		t.Fatalf("response request ID = %q", got)
	}
	if len(*log.entries) != 1 {
		t.Fatalf("log entries = %d, want 1", len(*log.entries))
	}
	fields := fieldMap((*log.entries)[0].fields)
	if fields["path"] != "/users/:id" || fields["status"] != http.StatusNoContent {
		t.Fatalf("unexpected fields: %#v", fields)
	}
	if _, leaked := fields["token"]; leaked {
		t.Fatal("query token was logged")
	}
}

func TestRecoverReturnsPanicAsError(t *testing.T) {
	e := echo.New()
	ctx := e.NewContext(httptest.NewRequest(http.MethodGet, "/panic", nil), httptest.NewRecorder())
	err := Recover(func(*echo.Context) error { panic("boom") })(ctx)
	if err == nil {
		t.Fatal("expected recovered panic error")
	}
}

func fieldMap(fields []logger.Field) map[string]any {
	result := make(map[string]any, len(fields))
	for _, field := range fields {
		result[field.Key] = field.Value
	}
	return result
}
