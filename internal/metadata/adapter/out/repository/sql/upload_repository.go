package sql

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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

func (r *UploadRepository) CreateUpload(ctx context.Context, upload domain.Upload) (domain.Upload, error) {
	query := `
		INSERT INTO uploads (upload_id, bucket, key, object_path, size, storage_node_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`

	db := GetDB(ctx, r.pool)

	uploadID, err := uuid.NewV7()
	if err != nil {
		return domain.Upload{}, fmt.Errorf("error: create new upload_id: %v", err)
	}

	if err := db.QueryRow(ctx, query, uploadID, upload.Bucket, upload.Key, upload.ObjectPath, upload.Size, upload.StorageNodeID).Scan(
		&upload.CreatedAt,
	); err != nil {
		return domain.Upload{}, err
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
		return fmt.Errorf("%w or already deleted: %s", domain.ErrUploadNotExists, uploadID.String())
	}

	return nil
}

func (r *UploadRepository) CommitUpload(ctx context.Context, uploadID uuid.UUID, checksum uint32) error {
	query := `
		WITH upload AS (
			DELETE FROM uploads
			WHERE upload_id = $1
			RETURNING bucket, key, object_path, size, storage_node_id
		)
		INSERT INTO objects (bucket, key, object_path, size, checksum, storage_node_id)
		SELECT bucket, key, object_path, size, $2, storage_node_id
		FROM upload
		ON CONFLICT (bucket, key)
		DO UPDATE SET 
			object_path = EXCLUDED.object_path,
			size = EXCLUDED.size,
			checksum = EXCLUDED.checksum,
			storage_node_id = EXCLUDED.storage_node_id
	`

	db := GetDB(ctx, r.pool)

	tag, err := db.Exec(ctx, query, uploadID, hash)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w or already commited: %s", domain.ErrUploadNotExists, uploadID)
	}

	return nil
}
