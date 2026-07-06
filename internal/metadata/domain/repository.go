package domain

import "context"

type MetadataRepository interface {
	GetObjectMetadata(ctx context.Context, bucket, key string) (ObjectMetadata, error)
	SaveObjectMetadata(ctx context.Context, meta ObjectMetadata) (ObjectMetadata, error)
	DeleteObjectMetadata(ctx context.Context, bucket, key string) error
}
