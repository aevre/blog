package logger

import (
	"bytes"
	"encoding/json"
	"runtime"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"blog-server/internal/platform/config"

	"go.uber.org/zap/zapcore"
)

func logThroughWrapper(log Logger) {
	log.WithCallerSkip(1).Warn("wrapped request")
}

func TestWithCallerSkip(t *testing.T) {
	core, entries := observer.New(zapcore.DebugLevel)
	log := &zapLogger{l: zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2))}
	_, file, line, _ := runtime.Caller(0)
	logThroughWrapper(log)
	entry := entries.All()[0]
	if entry.Caller.File != file || entry.Caller.Line != line+1 {
		t.Fatalf("caller = %s, want %s:%d", entry.Caller, file, line+1)
	}
}

func TestErrorOriginOverridesLogCaller(t *testing.T) {
	core, entries := observer.New(zapcore.DebugLevel)
	log := &zapLogger{l: zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2))}
	origin := runtime.Frame{File: "/app/handler/post.go", Line: 104, Function: "CreatePost"}
	logThroughWrapper(log.With(Any("error_origin", origin)))
	entry := entries.All()[0]
	if entry.Caller.File != origin.File || entry.Caller.Line != origin.Line {
		t.Fatalf("caller = %s, want error origin", entry.Caller)
	}
	if _, exists := entry.ContextMap()["error_origin"]; exists {
		t.Fatal("internal origin metadata leaked into log fields")
	}
}

func TestStackFormatting(t *testing.T) {
	for _, format := range []string{"console", "json"} {
		t.Run(format, func(t *testing.T) {
			var output bytes.Buffer
			encoder, level := buildEncoder(&config.Config{App: config.AppConfig{Environment: config.EnvDev}, Log: config.LogConfig{Format: format, Level: "debug"}})
			log := &zapLogger{l: zap.New(zapcore.NewCore(encoder, zapcore.AddSync(&output), level))}
			stack := "service.Refresh\n\tservice/auth.go:123"
			log.With(String("stack", stack)).WithCallerSkip(1).Warn("request failed", Int("status", 401))
			if format == "console" {
				if !strings.Contains(output.String(), "\n"+stack) || strings.Contains(output.String(), `\n`) {
					t.Fatalf("stack is not multiline: %q", output.String())
				}
			} else {
				var entry map[string]any
				if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
					t.Fatal(err)
				}
				if entry["stack"] != stack || strings.Count(output.String(), "\n") != 1 {
					t.Fatalf("unexpected JSON stack: %q", output.String())
				}
			}
		})
	}
}

func TestConfiguredLogLevel(t *testing.T) {
	log, ok := NewLogger(&config.Config{Log: config.LogConfig{Level: "warn", Format: "json"}}).(*zapLogger)
	if !ok {
		t.Fatal("unexpected logger implementation")
	}
	if log.l.Core().Enabled(zapcore.InfoLevel) {
		t.Fatal("info level should be disabled")
	}
	if !log.l.Core().Enabled(zapcore.WarnLevel) {
		t.Fatal("warn level should be enabled")
	}
}
