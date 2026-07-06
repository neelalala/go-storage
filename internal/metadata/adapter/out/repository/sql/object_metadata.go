package sql

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neelalala/go-storage/internal/metadata/domain"
)

type ObjectMetadataRepository struct {
	pool *pgxpool.Pool
}

func NewObjectMetadataRepository(pool *pgxpool.Pool) *ObjectMetadataRepository {
	return &ObjectMetadataRepository{
		pool: pool,
	}
}

func (r *ObjectMetadataRepository) GetObjectMetadata(ctx context.Context, bucket, key string) (domain.ObjectMetadata, error) {
	query := `
	SELECT bucket, key, size, checksum,
		created_at, updated_at, storage_node_id
	FROM object_metadata
	WHERE
		bucket = $1 AND key = $2
	`

	var meta domain.ObjectMetadata
	err := r.pool.QueryRow(ctx, query, bucket, key).Scan(
		&meta.Bucket,
		&meta.Key,
		&meta.Size,
		&meta.Checksum,
		&meta.CreatedAt,
		&meta.UpdatedAt,
		&meta.StorageNodeID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ObjectMetadata{}, fmt.Errorf("%w %s/%s", domain.ErrObjectNotFound, bucket, key)
		}
		return domain.ObjectMetadata{}, fmt.Errorf("error get object metadata: %v", err)
	}

	return meta, nil
}

func (r *ObjectMetadataRepository) SaveObjectMetadata(ctx context.Context, meta domain.ObjectMetadata) (domain.ObjectMetadata, error) {
	query := `
		INSERT INTO object_metadata (bucket, key, size, checksum, storage_node_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (bucket, key) DO UPDATE
		SET 
			bucket = EXCLUDED.bucket
		RETURNING bucket, key, size, checksum, created_at, updated_at, storage_node_id
	`

	var saved domain.ObjectMetadata
	err := r.pool.QueryRow(ctx, query,
		meta.Bucket,
		meta.Key,
		meta.Size,
		meta.Checksum,
		meta.StorageNodeID,
	).Scan(
		&saved.Bucket,
		&saved.Key,
		&saved.Size,
		&saved.Checksum,
		&saved.CreatedAt,
		&saved.UpdatedAt,
		&saved.StorageNodeID,
	)
	if err != nil {
		return domain.ObjectMetadata{}, fmt.Errorf("error saving object %s/%s metadata: %v", meta.Bucket, meta.Key, err)
	}

	return saved, nil
}

func (r *ObjectMetadataRepository) DeleteObjectMetadata(ctx context.Context, bucket, key string) error {
	query := `
		DELETE FROM object_metadata
		WHERE
			bucket = $1 AND key = $2
	`

	tag, err := r.pool.Exec(ctx, query, bucket, key)
	if err != nil {
		return fmt.Errorf("error deleting object %s/%s metadata: %v", bucket, key, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("error deleting object %s/%s metadata: %w", bucket, key, domain.ErrObjectNotFound)
	}

	return nil
}
