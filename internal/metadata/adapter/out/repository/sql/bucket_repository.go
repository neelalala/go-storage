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

var _ domain.BucketRepository = (*BucketRepository)(nil)

type BucketRepository struct {
	pool *pgxpool.Pool
}

func NewBucketRepository(pool *pgxpool.Pool) *BucketRepository {
	return &BucketRepository{
		pool: pool,
	}
}

func (r *BucketRepository) GetBuckets(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Bucket, error) {
	query := `
		SELECT name, created_at
		FROM buckets
		WHERE owner_id = $1
		ORDER BY name, created_at
		LIMIT $2 OFFSET $3
	`

	db := GetDB(ctx, r.pool)

	rows, err := db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	buckets := make([]domain.Bucket, 0, limit)
	for rows.Next() {
		var bucket domain.Bucket
		err := rows.Scan(
			&bucket.Name,
			&bucket.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		bucket.OwnerID = userID

		buckets = append(buckets, bucket)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return buckets, nil
}

func (r *BucketRepository) CreateBucket(ctx context.Context, userID uuid.UUID, name string) (domain.Bucket, error) {
	query := `
		INSERT INTO buckets (name, owner_id)
		VALUES ($1, $2)
		RETURNING created_at
	`

	db := GetDB(ctx, r.pool)

	var bucket domain.Bucket
	err := db.QueryRow(ctx, query, name, userID).Scan(
		&bucket.CreatedAt,
	)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return domain.Bucket{}, fmt.Errorf("%w: %s", domain.ErrBucketAlreadyExists, name)
			}
		}
		return domain.Bucket{}, err
	}
	bucket.Name = name
	bucket.OwnerID = userID

	return bucket, nil
}

func (r *BucketRepository) DeleteBucket(ctx context.Context, name string) error {
	query := `
		DELETE FROM buckets
		WHERE name = $1
	`

	db := GetDB(ctx, r.pool)

	tag, err := db.Exec(ctx, query, name)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == pgerrcode.ForeignKeyViolation {
				return fmt.Errorf("%w: %s", domain.ErrBucketNotEmpty, name)
			}
		}
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", domain.ErrBucketNotExists, name)
	}

	return nil
}

func (r *BucketRepository) GetBucket(ctx context.Context, name string) (domain.Bucket, error) {
	query := `
		SELECT created_at, owner_id
		FROM buckets
		WHERE name = $1
	`

	db := GetDB(ctx, r.pool)

	var bucket domain.Bucket
	err := db.QueryRow(ctx, query, name).Scan(
		&bucket.CreatedAt,
		&bucket.OwnerID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Bucket{}, fmt.Errorf("%w: %s", domain.ErrBucketNotExists, name)
		}
		return domain.Bucket{}, err
	}
	bucket.Name = name

	return bucket, nil
}
