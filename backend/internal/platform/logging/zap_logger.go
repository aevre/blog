package logger

import (
	"context"
	"errors"
	"os"
	"runtime"
	"syscall"
	"time"

	"blog-server/internal/platform/config"
	"blog-server/internal/shared/contextx"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapLogger is a Logger implementation backed by Uber Zap.
// It wraps zap.Logger to provide a simplified structured logging interface.
type zapLogger struct {
	l      *zap.Logger
	stack  string
	origin *runtime.Frame
}

// Debug logs a message at debug level with structured fields.
func (z *zapLogger) Debug(msg string, fields ...Field) { z.log(zap.DebugLevel, msg, fields...) }

// Info logs a message at info level with structured fields.
func (z *zapLogger) Info(msg string, fields ...Field) { z.log(zap.InfoLevel, msg, fields...) }

// Warn logs a message at warn level with structured fields.
func (z *zapLogger) Warn(msg string, fields ...Field) { z.log(zap.WarnLevel, msg, fields...) }

// Error logs a message at error level with structured fields.
func (z *zapLogger) Error(msg string, fields ...Field) { z.log(zap.ErrorLevel, msg, fields...) }

// Sync flushes any buffered log entries.
//
// It is safe to call multiple times. It ignores the common "invalid argument"
// error returned by stderr/stdout sync on some platforms.
func (z *zapLogger) Sync() error {
	err := z.l.Sync()
	if errors.Is(err, syscall.EINVAL) {
		return nil
	}
	return err
}

// With returns a new Logger instance enriched with the provided structured fields.
func (z *zapLogger) With(fields ...Field) Logger {
	child := *z
	regular := make([]Field, 0, len(fields))
	for _, field := range fields {
		if stack, ok := field.Value.(string); field.Key == "stack" && ok {
			child.stack = stack
		} else if origin, ok := field.Value.(runtime.Frame); field.Key == "error_origin" && ok {
			child.origin = &origin
		} else {
			regular = append(regular, field)
		}
	}
	child.l = z.l.With(convert(regular)...)
	return &child
}

// WithCallerSkip returns a logger that skips additional wrapper frames.
func (z *zapLogger) WithCallerSkip(skip int) Logger {
	child := *z
	child.l = z.l.WithOptions(zap.AddCallerSkip(skip))
	return &child
}

// WithContext extracts supported values from context and attaches them as fields.
//
// Currently supported keys:
//   - request_id
//   - trace_id
//
// If no known values exist in the context, the original logger is returned.
func (z *zapLogger) WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return z
	}

	if requestID := contextx.GetRequestID(ctx); requestID != "" {
		return z.With(String("request_id", requestID))
	}
	return z
}

// NewLogger creates a new Logger instance based on application configuration.
//
// It configures zap encoder, log level, and output format depending on
// the current environment (development or production).
func NewLogger(cfg *config.Config) Logger {
	enc, level := buildEncoder(cfg)

	infoPriority := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l >= level && l < zapcore.ErrorLevel
	})

	errorPriority := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l >= zapcore.ErrorLevel
	})

	infoCore := zapcore.NewCore(
		enc,
		zapcore.AddSync(os.Stdout),
		infoPriority,
	)

	errorCore := zapcore.NewCore(
		enc,
		zapcore.AddSync(os.Stderr),
		errorPriority,
	)

	core := zapcore.NewTee(infoCore, errorCore)

	l := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2))

	return &zapLogger{l: l}
}

// buildEncoder constructs zap encoder and log level based on environment.
//
// In production:
//   - Uses JSON encoder
//   - ISO8601 time format
//   - Info level and above
//
// In development:
//   - Uses colored console encoder
//   - Human-readable timestamps
//   - Debug level and above
func buildEncoder(cfg *config.Config) (zapcore.Encoder, zapcore.Level) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Log.Level)); err != nil {
		level = zap.InfoLevel
	}
	if cfg.Log.Format == "json" || cfg.App.IsProd() {
		ec := zap.NewProductionEncoderConfig()
		ec.StacktraceKey = "stack"
		ec.EncodeTime = zapcore.ISO8601TimeEncoder
		return zapcore.NewJSONEncoder(ec), level
	}

	ec := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stack",

		EncodeTime: func(t time.Time, pae zapcore.PrimitiveArrayEncoder) {
			pae.AppendString(t.Format("2006-01-02 15:04:05"))
		},

		// EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeLevel: func(l zapcore.Level, pae zapcore.PrimitiveArrayEncoder) {
			pae.AppendString("[" + l.CapitalString() + "]")
		},
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	return zapcore.NewConsoleEncoder(ec), level
}

// convert transforms internal Field types into zap.Field slice.
//
// It handles special cases such as error fields (zap.Error vs named errors)
// and falls back to zap.Any for generic values.
func convert(fields []Field) []zap.Field {
	if len(fields) == 0 {
		return nil
	}

	out := make([]zap.Field, 0, len(fields))

	for _, f := range fields {
		switch v := f.Value.(type) {
		case error:
			if f.Key == "error" {
				out = append(out, zap.Error(v))
			} else {
				out = append(out, zap.NamedError(f.Key, v))
			}
		default:
			out = append(out, zap.Any(f.Key, v))
		}
	}

	return out
}

// log writes a log entry at the specified level with structured fields.
func (z *zapLogger) log(level zapcore.Level, msg string, fields ...Field) {
	if entry := z.l.Check(level, msg); entry != nil {
		if z.origin != nil {
			entry.Caller = zapcore.EntryCaller{Defined: true, PC: z.origin.PC, File: z.origin.File, Line: z.origin.Line, Function: z.origin.Function}
		}
		if z.stack != "" {
			entry.Stack = z.stack
		}
		entry.Write(convert(fields)...)
	}
}

// Fatal logs a message at fatal level and then exits the application.
func (z *zapLogger) Fatal(msg string, fields ...Field) {
	z.l.Fatal(msg, convert(fields)...)
}

// Panic logs a message at panic level and then panics.
func (z *zapLogger) Panic(msg string, fields ...Field) {
	z.l.Panic(msg, convert(fields)...)
}
