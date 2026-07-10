package sql

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neelalala/go-storage/internal/metadata/domain"
)

type UploadRepository struct {
	pool *pgxpool.Pool
}

func NewUploadRepository(pool *pgxpool.Pool) *UploadRepository {
	return &UploadRepository{
		pool: pool,
	}
}

func (r *UploadRepository) CreateUpload(ctx context.Context, upload domain.Upload) (*domain.Upload, error) {
	query := `
		INSERT INTO uploads ($1, bucket, key, object_path, storage_node_id)
		VALUES ($2, $3, $4, $5)
		RETURNING upload_id, created_at
	`

	db := GetDB(ctx, r.pool)

	uploadID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("error: create new upload_id: %v", err)
	}

	if err := db.QueryRow(ctx, query, uploadID, upload.Bucket, upload.Key, upload.ObjectPath, upload.StorageNodeID).Scan(
		&upload.UploadID,
		&upload.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &upload, nil
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
		return fmt.Errorf("error: upload %s not exists", uploadID.String())
	}

	return nil
}

func (r *UploadRepository) CommitUpload(ctx context.Context, uploadID uuid.UUID, size uint64, checksum uint32) error {
	query := `
		WITH upload AS (
			DELETE FROM uploads
			WHERE upload_id = $1
			RETURNING bucket, key, object_path, storage_node_id
		)
		INSERT INTO objects (bucket, key, object_path, size, checksum, storage_node_id)
		SELECT bucket, key, object_path, $2, $3, storage_node_id
		FROM upload
		ON CONFLICT (bucket, key)
		DO UPDATE SET 
			object_path = EXCLUDED.object_path,
			size = EXCLUDED.size,
			checksum = EXCLUDED.checksum,
			storage_node_id = EXCLUDED.storage_node_id
	`

	db := GetDB(ctx, r.pool)

	tag, err := db.Exec(ctx, query, uploadID, size, checksum)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return errors.New("upload not found or was commited or aborted")
	}

	return nil
}
