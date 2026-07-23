package sql

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neelalala/go-storage/internal/metadata/domain"
)

var _ domain.UploadRepository = (*UploadRepository)(nil)

type UploadRepository struct {
	pool *pgxpool.Pool
}

func NewUploadRepository(pool *pgxpool.Pool) *UploadRepository {
	return &UploadRepository{
		pool: pool,
	}
}

func (r *UploadRepository) GetUpload(ctx context.Context, uploadID uuid.UUID) (domain.Upload, error) {
	query := `
		SELECT bucket, key, object_path, size, storage_node_id, created_at, 
			content_type, system_metadata, user_metadata, owner_id
		FROM uploads
		WHERE upload_id = $1
	`

	db := GetDB(ctx, r.pool)

	var upload domain.Upload
	if err := db.QueryRow(ctx, query, uploadID).Scan(
		&upload.Bucket,
		&upload.Key,
		&upload.ObjectPath,
		&upload.Size,
		&upload.StorageNodeID,
		&upload.CreatedAt,
		&upload.ContentType,
		&upload.SystemMetadata,
		&upload.UserMetadata,
		&upload.OwnerID,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Upload{}, fmt.Errorf("error getting upload: %w", domain.ErrUploadNotExists)
		}
		return domain.Upload{}, fmt.Errorf("error getting upload: %v", err)
	}
	upload.ID = uploadID

	return upload, nil
}

func (r *UploadRepository) CreateUpload(ctx context.Context, upload domain.Upload) (domain.Upload, error) {
	query := `
		INSERT INTO uploads 
		    (upload_id, bucket, key, object_path, size, storage_node_id,
		     content_type, system_metadata, user_metadata, owner_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at
	`

	db := GetDB(ctx, r.pool)

	uploadID, err := uuid.NewV7()
	if err != nil {
		return domain.Upload{}, fmt.Errorf("error creating new upload_id: %v", err)
	}

	if err := db.QueryRow(ctx, query,
		uploadID,
		upload.Bucket,
		upload.Key,
		upload.ObjectPath,
		upload.Size,
		upload.StorageNodeID,
		upload.ContentType,
		upload.SystemMetadata,
		upload.UserMetadata,
		upload.OwnerID,
	).Scan(
		&upload.CreatedAt,
	); err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == pgerrcode.ForeignKeyViolation {
				return domain.Upload{}, fmt.Errorf("error creating new uplaod: %w", domain.ErrBucketNotExists)
			}
		}
		return domain.Upload{}, fmt.Errorf("error creating new uplaod: %v", err)
	}

	upload.ID = uploadID

	return upload, nil
}

func (r *UploadRepository) DeleteUpload(ctx context.Context, uploadID uuid.UUID) error {
	query := `
		DELETE FROM uploads
		WHERE upload_id = $1
	`

	db := GetDB(ctx, r.pool)

	tag, err := db.Exec(ctx, query, uploadID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("error deleting upload: %w", domain.ErrUploadNotExists)
	}

	return nil
}

func (r *UploadRepository) CommitUpload(ctx context.Context, uploadID uuid.UUID, hash string) error {
	query := `
		WITH upload AS (
			DELETE FROM uploads
			WHERE upload_id = $1
			RETURNING bucket, key, object_path, size, storage_node_id, 
				content_type, system_metadata, user_metadata, owner_id
		)
		INSERT INTO objects 
			(bucket, key, object_path, size, hash, storage_node_id, 
			content_type, system_metadata, user_metadata, owner_id)
		SELECT bucket, key, object_path, size, $2, storage_node_id,
			content_type, system_metadata, user_metadata, owner_id
		FROM upload
		ON CONFLICT (bucket, key)
		DO UPDATE SET 
			object_path = EXCLUDED.object_path,
			size = EXCLUDED.size,
			hash = EXCLUDED.hash,
			storage_node_id = EXCLUDED.storage_node_id,
			content_type = EXCLUDED.content_type,
			system_metadata = EXCLUDED.system_metadata,
			user_metadata = EXCLUDED.user_metadata, 
	`

	db := GetDB(ctx, r.pool)

	tag, err := db.Exec(ctx, query, uploadID, hash)
	if err != nil {
		return fmt.Errorf("error commiting upload: %v", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("error commiting upload: %w", domain.ErrUploadNotExists)
	}

	return nil
}
