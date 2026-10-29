// Package contextx stores typed request-scoped values in contexts.
package contextx

import "context"

const requestKey contextKey = "request_id"

func SetRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestKey, id)
}

func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(requestKey).(string)
	return id
}
