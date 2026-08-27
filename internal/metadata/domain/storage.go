package domain

import (
	"github.com/google/uuid"
)

type StorageNode struct {
	ID      uuid.UUID
	Address string
}
