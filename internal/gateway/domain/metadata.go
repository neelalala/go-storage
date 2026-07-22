package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MetadataService interface {
	ListBuckets(ctx context.Context, userID uuid.UUID, limit, offset int) ([]BucketMetadata, error)
	CreateBucket(ctx context.Context, userID uuid.UUID, name string) (BucketMetadata, error)
	DeleteBucket(ctx context.Context, userID uuid.UUID, name string) error
	InitUpload(ctx context.Context, userID uuid.UUID, bucket, key string, size uint64, contentType string, systemMetadata, userMetadata map[string]string) (Upload, StorageNode, error)
	CommitUpload(ctx context.Context, userID uuid.UUID, uploadID uuid.UUID, etag string) error
	AbortUpload(ctx context.Context, userID uuid.UUID, uploadID uuid.UUID) error
	GetObject(ctx context.Context, userID uuid.UUID, bucket, key string) (ObjectMetadata, StorageNode, error)
	ListObjects(ctx context.Context, userID uuid.UUID, bucket, prefix, delimiter string, limit, offset int) ([]ObjectMetadata, error)
	DeleteObject(ctx context.Context, userID uuid.UUID, bucket, key string) error
}

type BucketMetadata struct {
	Name      string
	OwnerID   uuid.UUID
	CreatedAt time.Time
}

type ObjectMetadata struct {
	Bucket         string
	Key            string
	ObjectPath     string
	Size           uint64
	StorageNodeID  uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ContentType    string
	ETag           string
	SystemMetadata map[string]string
	UserMetadata   map[string]string
	OwnerID        uuid.UUID
}

type Upload struct {
	UploadID       uuid.UUID
	Bucket         string
	Key            string
	ObjectPath     string
	Size           uint64
	StorageNodeID  uuid.UUID
	CreatedAt      time.Time
	ContentType    string
	SystemMetadata map[string]string
	UserMetadata   map[string]string
	OwnerID        uuid.UUID
}
