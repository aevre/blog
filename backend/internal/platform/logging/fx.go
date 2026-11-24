package logger

import (
	"strings"
	"time"

	"go.uber.org/fx/fxevent"
)

type FxLogger struct {
	log Logger
}

func NewFxLogger(logger Logger) fxevent.Logger {
	return &FxLogger{log: logger}
}

func (l *FxLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.log.Error("on start failed", append(hookFields(e.FunctionName, e.CallerName, e.Runtime), Err(e.Err))...)
		}
	case *fxevent.OnStopExecuting:
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.log.Error("on stop failed", append(hookFields(e.FunctionName, e.CallerName, e.Runtime), Err(e.Err))...)
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			l.log.Error("supply failed", fxFields(
				e.ModuleName,
				String("type", e.TypeName),
				Any("stacktrace", e.StackTrace),
				Any("moduletrace", e.ModuleTrace),
				Err(e.Err),
			)...)
		}
	case *fxevent.Provided:
		if e.Err != nil {
			l.log.Error("provide failed", fxFields(
				e.ModuleName,
				String("constructor", e.ConstructorName),
				Any("types", e.OutputTypeNames),
				Any("stacktrace", e.StackTrace),
				Any("moduletrace", e.ModuleTrace),
				Err(e.Err),
			)...)
		}
	case *fxevent.Replaced:
		if e.Err != nil {
			l.log.Error("replace failed", fxFields(
				e.ModuleName,
				Any("types", e.OutputTypeNames),
				Any("stacktrace", e.StackTrace),
				Any("moduletrace", e.ModuleTrace),
				Err(e.Err),
			)...)
		}
	case *fxevent.Decorated:
		if e.Err != nil {
			l.log.Error("decorate failed", fxFields(
				e.ModuleName,
				String("decorator", e.DecoratorName),
				Any("types", e.OutputTypeNames),
				Any("stacktrace", e.StackTrace),
				Any("moduletrace", e.ModuleTrace),
				Err(e.Err),
			)...)
		}
	case *fxevent.BeforeRun:
	case *fxevent.Run:
		if e.Err != nil {
			l.log.Error("run failed", fxFields(
				e.ModuleName,
				String("name", e.Name),
				String("kind", e.Kind),
				Duration("runtime", e.Runtime),
				Err(e.Err),
			)...)
		}
	case *fxevent.Invoking:
	case *fxevent.Invoked:
		if e.Err != nil {
			l.log.Error("invoke failed", fxFields(e.ModuleName, String("function", e.FunctionName), String("stack", e.Trace), Err(e.Err))...)
		}
	case *fxevent.Stopping:
		if e.Signal != nil {
			l.log.Debug("received signal", String("signal", strings.ToUpper(e.Signal.String())))
		}
	case *fxevent.Stopped:
		if e.Err != nil {
			l.log.Error("stop failed", Err(e.Err))
		}
	case *fxevent.RollingBack:
		l.log.Error("start failed, rolling back", Err(e.StartErr))
	case *fxevent.RolledBack:
		if e.Err != nil {
			l.log.Error("rollback failed", Err(e.Err))
		}
	case *fxevent.Started:
		if e.Err != nil {
			l.log.Error("start failed", Err(e.Err))
		} else {
			l.log.Debug("started")
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			l.log.Error("custom logger initialization failed", Err(e.Err))
		}
	}
}

func hookFields(function, caller string, runtime time.Duration) []Field {
	return []Field{String("function", function), String("caller", caller), Duration("runtime", runtime)}
}

func fxFields(module string, fields ...Field) []Field {
	if module != "" {
		fields = append(fields, String("module", module))
	}
	return fields
}
