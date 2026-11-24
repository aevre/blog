package logger

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestFxLoggerSuppressesRoutineEvents(t *testing.T) {
	core, entries := observer.New(zapcore.DebugLevel)
	fxlog := NewFxLogger(&zapLogger{l: zap.New(core)})

	fxlog.LogEvent(&fxevent.Provided{
		ConstructorName: "NewService",
		OutputTypeNames: []string{"*Service", "io.Closer"},
		ModuleName:      "service",
		Private:         true,
	})
	fxlog.LogEvent(&fxevent.Provided{ConstructorName: "NewEmpty"})
	fxlog.LogEvent(&fxevent.Supplied{TypeName: "*Config"})
	fxlog.LogEvent(&fxevent.Run{Name: "NewService", Kind: "provide"})
	fxlog.LogEvent(&fxevent.Invoked{FunctionName: "registerRoutes"})

	if got := entries.All(); len(got) != 0 {
		t.Fatalf("logged %d routine Fx events, want none", len(got))
	}
}

func TestFxLoggerLogsHookFailure(t *testing.T) {
	core, entries := observer.New(zapcore.DebugLevel)
	fxlog := NewFxLogger(&zapLogger{l: zap.New(core)})
	err := errors.New("startup failed")

	fxlog.LogEvent(&fxevent.OnStartExecuted{
		FunctionName: "startService",
		CallerName:   "registerService",
		Runtime:      2 * time.Second,
		Err:          err,
	})

	got := entries.All()
	if len(got) != 1 {
		t.Fatalf("logged %d events, want 1", len(got))
	}
	if got[0].Level != zapcore.ErrorLevel || got[0].Message != "on start failed" {
		t.Fatalf("entry = %s %q, want error on start failed", got[0].Level, got[0].Message)
	}
	fields := got[0].ContextMap()
	if fields["function"] != "startService" || fields["caller"] != "registerService" || fields["runtime"] != 2*time.Second {
		t.Fatalf("unexpected fields: %#v", fields)
	}
	if fields["error"] != err.Error() {
		t.Fatalf("error = %#v, want %q", fields["error"], err)
	}
}
