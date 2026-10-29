package middleware

import (
	"io"
	"net/http"
	"sync"

	"github.com/labstack/echo/v5"
)

// BodyLimitConfig defines the config for BodyLimitWithConfig middleware.
type BodyLimitConfig struct {
	Skipper    Skipper
	LimitBytes int64
}

func BodyLimit(limitBytes int64) echo.MiddlewareFunc {
	return BodyLimitWithConfig(BodyLimitConfig{LimitBytes: limitBytes})
}

func BodyLimitWithConfig(config BodyLimitConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}

	limit := config.LimitBytes

	pool := sync.Pool{
		New: func() any {
			return &limitedReader{limit: limit}
		},
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			req := c.Request()

			// Based on content length
			if req.ContentLength > config.LimitBytes {
				return echo.ErrStatusRequestEntityTooLarge
			}

			// Based on content read
			r, ok := pool.Get().(*limitedReader)
			if !ok {
				return echo.NewHTTPError(http.StatusInternalServerError, "invalid pool object")
			}
			r.Reset(req.Body)
			req.Body = r
			defer func() {
				_ = r.Close()
			}()

			return next(c)
		}
	}
}

type limitedReader struct {
	pool   *sync.Pool
	reader io.ReadCloser
	limit  int64
	read   int64
}

func (r *limitedReader) Read(b []byte) (n int, err error) {
	n, err = r.reader.Read(b)
	r.read += int64(n)
	if r.read > r.limit {
		return n, echo.ErrStatusRequestEntityTooLarge
	}
	return
}

func (r *limitedReader) Close() error {
	err := r.reader.Close()
	r.reader = nil
	if r.pool != nil {
		r.pool.Put(r)
	}
	return err
}

func (r *limitedReader) Reset(reader io.ReadCloser) {
	r.reader = reader
	r.read = 0
}
