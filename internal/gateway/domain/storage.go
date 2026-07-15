package domain

import (
	"context"

	"github.com/google/uuid"
)

type Storage interface {
	SaveObject(ctx context.Context, object Object) (uint32, error)
	GetObject(ctx context.Context, name string) (Object, error)
	DeleteObject(ctx context.Context, name string) error
}

type StorageNode struct {
	ID      uuid.UUID
	Address string
}

type StorageNodeManager interface {
	GetStorage(address string) (Storage, error)
	Invalidate(address string)
}
