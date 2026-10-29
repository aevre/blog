// Package logger provides logging capabilities
package logger

import (
	"context"
	"time"
)

// Logger defines a structured logging interface with leveled logging support
// and context propagation capabilities.
//
// Implementations should support structured key-value logging.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)

	Fatal(msg string, fields ...Field)
	Panic(msg string, fields ...Field)

	// With returns a new Logger instance with additional structured fields
	// attached to all subsequent log entries.
	With(fields ...Field) Logger

	// WithCallerSkip skips additional logging wrapper frames when reporting the caller.
	WithCallerSkip(skip int) Logger

	// WithContext attaches a context to the logger, typically used for
	// request-scoped logging (e.g., trace ID, request ID).
	WithContext(ctx context.Context) Logger

	// Sync flushes any buffered log entries.
	Sync() error
}

// Field represents a structured logging key-value pair.
type Field struct {
	Key   string
	Value any
}

// Any creates a generic structured logging field.
func Any(key string, value any) Field {
	return Field{Key: key, Value: value}
}

// Err creates a structured field for an error value using the standard key "error".
func Err(err error) Field {
	return Field{Key: "error", Value: err}
}

// String creates a string-type structured field.
func String(key, v string) Field { return Field{Key: key, Value: v} }

// Int creates an int-type structured field.
func Int(key string, v int) Field { return Field{Key: key, Value: v} }

// Int64 creates an int64-type structured field.
func Int64(key string, v int64) Field { return Field{Key: key, Value: v} }

// Bool creates a bool-type structured field.
func Bool(key string, v bool) Field { return Field{Key: key, Value: v} }

// Float64 creates a float64-type structured field.
func Float64(key string, v float64) Field { return Field{Key: key, Value: v} }

// Duration creates a time.Duration structured field.
func Duration(key string, v time.Duration) Field { return Field{Key: key, Value: v} }

// Time creates a time.Time structured field.
func Time(key string, v time.Time) Field { return Field{Key: key, Value: v} }
