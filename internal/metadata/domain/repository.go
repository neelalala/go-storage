package domain

import (
	"context"

	"github.com/google/uuid"
)

type Transactor interface {
	WithinTransaction(context.Context, func(context.Context) error) error
}

type BucketRepository interface {
	// GetBuckets may return next domain errors: ErrAccessDenied
	GetBuckets(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Bucket, error)
	// CreateBucket may return next domain errors: ErrBucketAlreadyExists
	CreateBucket(ctx context.Context, userID uuid.UUID, name string) (Bucket, error)
	// DeleteBucket may return next domain errors: ErrBucketNotEmpty, ErrBucketNotExists
	DeleteBucket(ctx context.Context, name string) error
	// GetBucket may return next domain errors: ErrBucketNotExists
	GetBucket(ctx context.Context, name string) (Bucket, error)
}

type UploadRepository interface {
	// GetUpload may return next domain errors: ErrUploadNotExists
	GetUpload(ctx context.Context, uploadID uuid.UUID) (Upload, error)
	// CreateUpload may return next domain errors: ErrBucketNotExists
	CreateUpload(ctx context.Context, upload Upload) (Upload, error)
	// DeleteUpload may return next domain errors: ErrUploadNotExists
	DeleteUpload(ctx context.Context, uploadID uuid.UUID) error
	// CommitUpload may return next domain errors: ErrUploadNotExists.
	// Deletes the entry in the uploads table and creates an entry in objects, atomically
	CommitUpload(ctx context.Context, uploadID uuid.UUID, hash string) error
}

type ObjectRepository interface {
	// GetObject may return next domain errors: ErrAccessDenied, ErrBucketNotExists, ErrObjectNotFound
	GetObject(ctx context.Context, userID uuid.UUID, bucket, key string) (Object, error)
	// SoftDeleteObject may return next domain errors: ErrBucketNotExists, ErrBucketNotEmpty
	SoftDeleteObject(ctx context.Context, bucket, key string) error
	// GetObjects returns slices with summarized len at maximum equal limit.
	// Slice of Object - objects with key that satisfy prefix and has no delimiter (except in prefix)
	// Slice of string - grouped objects that have prefix but also delimiter in their key
	GetObjects(ctx context.Context, bucket, path, delimiter string, limit, offset int) ([]Object, []string, error)
}

type GCRepository interface {
	GetPendingGCTasks(ctx context.Context, limit int) ([]GCTask, error)
	CompleteGCTask(ctx context.Context, deletionID int64) error
	IncrementGCTaskAttempts(ctx context.Context, deletionID int64) error
}
