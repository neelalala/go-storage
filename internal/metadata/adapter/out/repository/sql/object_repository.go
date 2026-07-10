package sql

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neelalala/go-storage/internal/metadata/domain"
)

type ObjectRepository struct {
	pool *pgxpool.Pool
}

func NewObjectRepository(pool *pgxpool.Pool) *ObjectRepository {
	return &ObjectRepository{
		pool: pool,
	}
}

func (r *ObjectRepository) GetObject(ctx context.Context, bucket, key string) (*domain.Object, error) {
	query := `
		SELECT bucket, key, object_path, size, checksum, storage_node_id, created_at, updated_at
		FROM objects
		WHERE bucket = $1 AND key = $2;
	`

	db := GetDB(ctx, r.pool)

	var object domain.Object
	if err := db.QueryRow(ctx, query, bucket, key).Scan(
		&object.Bucket,
		&object.Key,
		&object.ObjectPath,
		&object.Size,
		&object.Checksum,
		&object.StorageNodeID,
		&object.CreatedAt,
		&object.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &object, nil
}

func (r *ObjectRepository) SoftDeleteObject(ctx context.Context, bucket, key string) (*domain.Object, error) {
	query := `
		WITH deleted AS (
			DELETE FROM objects
			WHERE bucket = $1 AND key = $2
			RETURNING bucket, key, object_path, size, checksum, storage_node_id, created_at, updated_at
		),
		WITH inserted AS (
			INSERT INTO gc_queue (object_path, storage_node_id)
			SELECT object_path, storage_node_id FROM deleted
		)
		SELECT bucket, key, object_path, size, checksum, storage_node_id, created_at, updated_at
		FROM deleted;
	`

	db := GetDB(ctx, r.pool)

	var object domain.Object
	if err := db.QueryRow(ctx, query, bucket, key).Scan(
		&object.Bucket,
		&object.Key,
		&object.ObjectPath,
		&object.Size,
		&object.Checksum,
		&object.StorageNodeID,
		&object.CreatedAt,
		&object.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s/%s", domain.ErrObjectNotFound, bucket, key)
		}
		return nil, err
	}

	return &object, nil
}
