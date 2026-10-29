// Package middleware
package middleware

import "github.com/labstack/echo/v5"

type Skipper func(c *echo.Context) bool

func DefaultSkipper(c *echo.Context) bool {
	return false
}
