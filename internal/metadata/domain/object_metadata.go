package domain

import (
	"time"

	"github.com/google/uuid"
)

type Object struct {
	Bucket        string
	Key           string
	ObjectPath    string
	Size          uint64
	Checksum      uint32
	StorageNodeID uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Upload struct {
	UploadID      uuid.UUID
	Bucket        string
	Key           string
	ObjectPath    string
	StorageNodeID uuid.UUID
	CreatedAt     time.Time
}

type Status string

const (
	StatusPending Status = "PENDING"
	StatusError   Status = "ERROR"
)

type GCTask struct {
	DeletionID    int64
	ObjectPath    string
	StorageNodeID uuid.UUID
	Status        Status
	Attempts      int
	CreatedAt     time.Time
}
