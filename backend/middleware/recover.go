package middleware

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/labstack/echo/v5"
)

type panicError struct {
	value any
	stack []byte
}

func (e *panicError) Error() string {
	return fmt.Sprintf("panic: %v\n%s", e.value, e.stack)
}

func Recover(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) (err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				if recovered == http.ErrAbortHandler {
					panic(recovered)
				}
				stack := make([]byte, 64<<10)
				stack = stack[:runtime.Stack(stack, false)]
				err = &panicError{value: recovered, stack: stack}
			}
		}()
		return next(c)
	}
}
