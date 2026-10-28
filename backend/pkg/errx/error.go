// Package errx
package errx

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// AppError is a low-level application error wrapper.
// It is NOT responsible for user-facing messages.
// It only carries machine-readable error metadata.
type AppError struct {
	Code  int
	Err   error
	stack []uintptr
}

// New creates a new AppError with a code and underlying error.
// It captures stack trace for debugging purposes.
func New(code int, err error) *AppError {
	if err == nil {
		return &AppError{
			Code:  code,
			Err:   nil,
			stack: captureStack(),
		}
	}

	// avoid double wrapping
	if ae, ok := errors.AsType[*AppError](err); ok {
		return ae
	}

	return &AppError{
		Code:  code,
		Err:   err,
		stack: captureStack(),
	}
}

// Error implements error interface.
// Keep it minimal for logs/debugging, not for user display.
func (e *AppError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("[%d]", e.Code)
	}
	return fmt.Sprintf("[%d] %v", e.Code, e.Err)
}

// Unwrap supports errors.Is / errors.As
func (e *AppError) Unwrap() error {
	return e.Err
}

// Stack returns raw stack trace frames.
func (e *AppError) Stack() []uintptr {
	return e.stack
}

// StackString formats stack trace for logging/debug only.
func (e *AppError) StackString() string {
	if len(e.stack) == 0 {
		return ""
	}

	frames := runtime.CallersFrames(e.stack)
	var out strings.Builder

	for {
		f, more := frames.Next()
		fmt.Fprintf(&out, "%s\n\t%s:%d\n", f.Function, f.File, f.Line)
		if !more {
			break
		}
	}

	return out.String()
}

// captureStack records current goroutine stack trace.
func captureStack() []uintptr {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(3, pcs) // skip: Callers + New + caller
	return pcs[:n]
}

// ToAppError converts any error into AppError.
func ToAppError(err error) *AppError {
	if err == nil {
		return nil
	}

	if ae, ok := errors.AsType[*AppError](err); ok {
		return ae
	}

	return New(CodeInternalError, err)
}

func MessageForCode(code int) string {
	switch code {
	case CodeNotFound:
		return "资源不存在"
	case CodeConflict:
		return "资源状态冲突"
	case CodeUnauthorized, CodeTokenExpired:
		return "未登录或登录已过期"
	case CodeForbidden:
		return "权限不足"
	case CodeInvalidParam, CodeValidationFailed:
		return "请求参数错误"
	case CodeExternalError:
		return "外部服务暂时不可用"
	default:
		return "系统错误，请稍后重试"
	}
}
