package domain

import "context"

type Store interface {
	Save(ctx context.Context, object Object) (string, error)
	Get(ctx context.Context, name string) ([]byte, error)
	Delete(ctx context.Context, name string) error
}
