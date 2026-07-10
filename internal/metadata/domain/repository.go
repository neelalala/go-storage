package domain

import (
	"context"

	"github.com/google/uuid"
)

type Transactor interface {
	WithinTransaction(context.Context, func(context.Context) error) error
}

type UploadRepository interface {
	CreateUpload(ctx context.Context, upload Upload) (*Upload, error)
	DeleteUpload(ctx context.Context, uploadID uuid.UUID) error
	CommitUpload(ctx context.Context, uploadID uuid.UUID, size uint64, checksum uint32) error
}

type ObjectRepository interface {
	GetObject(ctx context.Context, bucket, key string) (*Object, error)
	SoftDeleteObject(ctx context.Context, bucket, key string) (*Object, error)
}

type GCRepository interface {
	GetPendingGCTasks(ctx context.Context, limit int) ([]*GCTask, error)
	CompleteGCTask(ctx context.Context, deletionID int64) error
	IncremetGCTaskAttempts(ctx context.Context, deletionID int64) error
}
