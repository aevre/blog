package handler

import (
	"errors"
	"net/http"
	"runtime"
	"time"

	"blog-server/config"
	"blog-server/logger"
	"blog-server/middleware"
	"blog-server/pkg/errx"
	"blog-server/response"

	"github.com/labstack/echo/v5"
)

func ErrorHandler(cfg *config.Config, log logger.Logger) echo.HTTPErrorHandler {
	return func(c *echo.Context, err error) {
		appErr := errx.ToAppError(err)
		httpCode := errx.MapToHTTPStatus(appErr.Code)
		var httpErr *echo.HTTPError
		if errors.As(err, &httpErr) {
			httpCode = httpErr.Code
			appErr = errx.New(codeForHTTPStatus(httpCode), err)
		}

		reqLogger := log
		if l, ok := c.Get(middleware.ContextKeyLogger).(logger.Logger); ok {
			reqLogger = l
		}
		duration := time.Duration(0)
		if start, ok := c.Get(middleware.ContextKeyStart).(time.Time); ok {
			duration = time.Since(start)
		}
		committed := false
		if resp, unwrapErr := echo.UnwrapResponse(c.Response()); unwrapErr == nil {
			committed = resp.Committed
			if committed {
				httpCode = resp.Status
			}
		}
		reqLogger = reqLogger.With(logger.Int("code", appErr.Code))
		if stack := appErr.Stack(); len(stack) > 0 {
			origin, _ := runtime.CallersFrames(stack).Next()
			reqLogger = reqLogger.With(logger.Any("error_origin", origin))
		}
		if cfg.App.IsDev() || httpCode >= http.StatusInternalServerError {
			reqLogger = reqLogger.With(logger.String("stack", appErr.StackString()))
		}
		middleware.LogRequest(reqLogger, c, httpCode, duration, err)

		if committed {
			return
		}
		if sendErr := c.JSON(httpCode, response.Error(appErr.Code, errx.MessageForCode(appErr.Code))); sendErr != nil {
			reqLogger.Error("failed to send error response", logger.Err(sendErr))
		}
	}
}

func codeForHTTPStatus(status int) int {
	switch status {
	case http.StatusBadRequest, http.StatusRequestEntityTooLarge:
		return errx.CodeInvalidParam
	case http.StatusUnauthorized:
		return errx.CodeUnauthorized
	case http.StatusForbidden:
		return errx.CodeForbidden
	case http.StatusNotFound:
		return errx.CodeNotFound
	case http.StatusConflict:
		return errx.CodeConflict
	default:
		return errx.CodeInternalError
	}
}
