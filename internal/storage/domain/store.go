package domain

import "context"

type Store interface {
	Save(ctx context.Context, object Object) (uint32, error)
	Get(ctx context.Context, name string) (Object, error)
	Delete(ctx context.Context, name string) error
}
