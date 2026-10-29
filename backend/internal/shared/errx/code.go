package errx

import "net/http"

const (
	CodeOK = 0

	CodeUnauthorized = 1001
	CodeForbidden    = 1002
	CodeTokenExpired = 1003

	CodeNotFound = 2001
	CodeConflict = 2002

	CodeInvalidParam     = 3001
	CodeValidationFailed = 3002

	CodeInternalError = 5000
	CodeExternalError = 5001
)

func MapToHTTPStatus(code int) int {
	switch code {

	case CodeOK:
		return http.StatusOK

	case CodeUnauthorized, CodeTokenExpired:
		return http.StatusUnauthorized

	case CodeForbidden:
		return http.StatusForbidden

	case CodeInvalidParam, CodeValidationFailed:
		return http.StatusBadRequest

	case CodeNotFound:
		return http.StatusNotFound

	case CodeConflict:
		return http.StatusConflict

	default:
		return http.StatusInternalServerError
	}
}
