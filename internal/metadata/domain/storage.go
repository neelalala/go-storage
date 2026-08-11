package domain

import (
	"context"

	"github.com/google/uuid"
)

type StorageNode struct {
	ID      uuid.UUID
	Address string
}

type NodeManager interface {
	NextNode(ctx context.Context) (StorageNode, error)
	GetNode(ctx context.Context, id uuid.UUID) (StorageNode, error)
}
