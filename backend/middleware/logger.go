package middleware

import (
	"net/http"
	"time"

	"blog-server/contextx"
	"blog-server/logger"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

const (
	ContextKeyLogger = "logger"
	ContextKeyStart  = "request_start"
)

func RequestLogger(log logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()
			reqID := c.Request().Header.Get(echo.HeaderXRequestID)
			if reqID == "" || len(reqID) > 128 {
				reqID = uuid.New().String()
			}
			c.Response().Header().Set(echo.HeaderXRequestID, reqID)

			reqLogger := log.With(logger.String("request_id", reqID))
			c.Set(ContextKeyLogger, reqLogger)
			c.Set(ContextKeyStart, start)
			c.SetRequest(c.Request().WithContext(contextx.SetRequestID(c.Request().Context(), reqID)))

			err := next(c)
			if err != nil {
				return err
			}

			status := http.StatusOK
			if resp, unwrapErr := echo.UnwrapResponse(c.Response()); unwrapErr == nil {
				status = resp.Status
			}
			LogRequest(reqLogger, c, status, time.Since(start), nil)
			return nil
		}
	}
}

func LogRequest(log logger.Logger, c *echo.Context, status int, duration time.Duration, err error) {
	log = log.WithCallerSkip(1)
	path := c.Path()
	if path == "" {
		path = c.Request().URL.Path
	}
	fields := []logger.Field{
		logger.String("method", c.Request().Method),
		logger.String("path", path),
		logger.Int("status", status),
		logger.Duration("duration", duration),
	}
	if err != nil {
		fields = append(fields, logger.Err(err))
	}

	switch {
	case status >= http.StatusInternalServerError:
		log.Error("http request completed", fields...)
	case status >= http.StatusBadRequest || duration > 5*time.Second:
		log.Warn("http request completed", fields...)
	default:
		log.Info("http request completed", fields...)
	}
}
