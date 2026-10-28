package logger

import (
	"testing"

	"blog-server/config"

	"go.uber.org/zap/zapcore"
)

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
