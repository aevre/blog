// Package httpx provides shared HTTP response envelopes and helpers.
package httpx

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

type Response[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type ErrorResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type Page[T any] struct {
	Total int `json:"total"`
	List  []T `json:"list"`
}

const CodeSuccess = 0

func OK(c *echo.Context, data any) error {
	return c.JSON(http.StatusOK, data)
}

func Success[T any](data T) *Response[T] {
	return &Response[T]{
		Code: CodeSuccess,
		Msg:  "success",
		Data: data,
	}
}

func SuccessWithMsg[T any](msg string, data T) *Response[T] {
	return &Response[T]{
		Code: CodeSuccess,
		Msg:  msg,
		Data: data,
	}
}

func Error(code int, msg string) *ErrorResponse {
	return &ErrorResponse{
		Code: code,
		Msg:  msg,
	}
}
