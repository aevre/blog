// Package storage
package storage

import (
	"context"
	"io"
)

type Storage interface {
	Upload(ctx context.Context, bucket, key string, body io.Reader, contentType string) error
	Download(ctx context.Context, bucket, key string) (io.ReadCloser, string, error)
	Delete(ctx context.Context, bucket, key string) error
	Copy(ctx context.Context, bucket, srcKey, dstKey string) error
	Exists(ctx context.Context, bucket, key string) (bool, error)
}
