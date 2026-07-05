package domain

import "context"

type Storage interface {
	SaveObject(ctx context.Context, name string, data []byte) error
	GetObject(ctx context.Context, name string) ([]byte, error)
	DeleteObject(ctx context.Context, name string) error
}
