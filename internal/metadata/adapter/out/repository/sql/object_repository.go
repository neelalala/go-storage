package sql

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neelalala/go-storage/internal/metadata/domain"
)

var _ domain.ObjectRepository = (*ObjectRepository)(nil)

type ObjectRepository struct {
	pool *pgxpool.Pool
}

func NewObjectRepository(pool *pgxpool.Pool) *ObjectRepository {
	return &ObjectRepository{
		pool: pool,
	}
}

func (r *ObjectRepository) GetObject(ctx context.Context, bucket, key string) (domain.Object, error) {
	query := `
		SELECT bucket, key, object_path, size, checksum, storage_node_id, created_at, updated_at
		FROM objects
		WHERE bucket = $1 AND key = $2;
	`

	db := GetDB(ctx, r.pool)

	var object domain.Object
	err := db.QueryRow(ctx, query, bucket, key).Scan(
		&object.Bucket,
		&object.Key,
		&object.ObjectPath,
		&object.Size,
		&object.Checksum,
		&object.StorageNodeID,
		&object.CreatedAt,
		&object.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Object{}, fmt.Errorf("%w: %s/%s", domain.ErrObjectNotFound, bucket, key)
		}
		return domain.Object{}, err
	}

	return object, nil
}

func (r *ObjectRepository) GetObjects(ctx context.Context, bucket, prefix, delimiter string, limit, offset int) ([]domain.Object, []string, error) {
	query := `
       SELECT 
           key AS item_name,
           false AS is_prefix,
           object_path, size, hash, storage_node_id, created_at, updated_at,
           content_type, system_metadata, user_metadata, owner_id
       FROM objects
       WHERE bucket = $1 
         AND key LIKE $2 || '%'
         AND ($3::text = '' OR strpos(substring(key from length($2::text) + 1), $3::text) = 0)

       UNION ALL

       SELECT DISTINCT 
           substring(key from 1 for length($2::text) + strpos(substring(key from length($2::text) + 1), $3::text) + length($3::text) - 1) AS item_name,
           true AS is_prefix,
           NULL::text, NULL::bigint, NULL::text, NULL::uuid, NULL::timestamp, NULL::timestamp,
           NULL::text, NULL::jsonb, NULL::jsonb, NULL::uuid
       FROM objects
       WHERE bucket = $1 
         AND $3::text != ''
         AND key LIKE $2 || '%'
         AND strpos(substring(key from length($2::text) + 1), $3::text) > 0

       ORDER BY item_name
       LIMIT $4 OFFSET $5;
    `

	db := GetDB(ctx, r.pool)

	rows, err := db.Query(ctx, query, bucket, prefix, delimiter, limit, offset)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	objects := make([]domain.Object, 0)
	commonPrefixes := make([]string, 0)

	for rows.Next() {
		var (
			itemName       string
			isPrefix       bool
			objectPath     *string
			size           *uint64
			hash           *string
			storageNodeID  *uuid.UUID
			createdAt      *time.Time
			updatedAt      *time.Time
			contentType    *string
			systemMetadata map[string]string
			userMetadata   map[string]string
			ownerID        *uuid.UUID
		)

		if err := rows.Scan(
			&itemName,
			&isPrefix,
			&objectPath,
			&size,
			&hash,
			&storageNodeID,
			&createdAt,
			&updatedAt,
			&contentType,
			&systemMetadata,
			&userMetadata,
			&ownerID,
		); err != nil {
			return nil, nil, err
		}

		if isPrefix {
			commonPrefixes = append(commonPrefixes, itemName)
		} else {
			objects = append(objects, domain.Object{
				Bucket:         bucket,
				Key:            itemName,
				ObjectPath:     *objectPath,
				Size:           *size,
				StorageNodeID:  *storageNodeID,
				CreatedAt:      *createdAt,
				UpdatedAt:      *updatedAt,
				ContentType:    *contentType,
				Hash:           *hash,
				SystemMetadata: systemMetadata,
				UserMetadata:   userMetadata,
				OwnerID:        *ownerID,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return objects, commonPrefixes, nil
}

func (r *ObjectRepository) SoftDeleteObject(ctx context.Context, bucket, key string) error {
	query := `
		WITH deleted AS (
			DELETE FROM objects
			WHERE bucket = $1 AND key = $2
			RETURNING object_path, storage_node_id
		)
		INSERT INTO gc_queue (object_path, storage_node_id)
		SELECT object_path, storage_node_id FROM deleted
	`

	db := GetDB(ctx, r.pool)

	tag, err := db.Exec(ctx, query, bucket, key)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s/%s", domain.ErrObjectNotFound, bucket, key)
	}

	return nil
}
