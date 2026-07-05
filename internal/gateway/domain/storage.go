package domain

import "context"

type Storage interface {
	SaveObject(ctx context.Context, object Object) error
	GetObject(ctx context.Context, name string) (Object, error)
	DeleteObject(ctx context.Context, name string) error
}
