package storage

import (
	"context"
	"io"
)

type UploadResult struct {
	URL string
	Key string
}

type Storage interface {
	Upload(ctx context.Context, file io.Reader, folder string) (UploadResult, error)
	Delete(ctx context.Context, key string) error
}
