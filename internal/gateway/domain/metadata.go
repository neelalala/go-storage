package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MetadataService interface {
	ListBuckets(ctx context.Context, limit, offset int) ([]BucketMetadata, error)
	CreateBucket(ctx context.Context, name string) (BucketMetadata, error)
	DeleteBucket(ctx context.Context, name string) error
	InitUpload(ctx context.Context, bucket, key string, size uint64) (Upload, StorageNode, error)
	CommitUpload(ctx context.Context, uploadID uuid.UUID, checksum uint32) error
	AbortUpload(ctx context.Context, uploadID uuid.UUID) error
	GetObject(ctx context.Context, bucket, key string) (ObjectMetadata, StorageNode, error)
	ListObjects(ctx context.Context, bucket, prefix, delimiter string, limit, offset int) ([]ObjectMetadata, error)
	DeleteObject(ctx context.Context, bucket, key string) error
}

type BucketMetadata struct {
	Name      string
	CreatedAt time.Time
}

type ObjectMetadata struct {
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
	Size          uint64
	StorageNodeID uuid.UUID
	CreatedAt     time.Time
}
