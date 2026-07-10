package sql

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neelalala/go-storage/internal/metadata/domain"
)

type GCRepository struct {
	pool *pgxpool.Pool
}

func NewGCRepository(pool *pgxpool.Pool) *GCRepository {
	return &GCRepository{
		pool: pool,
	}
}

func (r *GCRepository) GetPendingGCTasks(ctx context.Context, limit int) ([]*domain.GCTask, error) {
	query := `
		SELECT deletion_id, object_path, storage_node_id, attempts, created_at
		FROM gc_queue
		WHERE status = $1
		ORDER BY created_at, attempts 
		LIMIT $2 
	`

	db := GetDB(ctx, r.pool)

	rows, err := db.Query(ctx, query, domain.StatusPending, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]*domain.GCTask, 0, limit)

	for rows.Next() {
		var task domain.GCTask

		err := rows.Scan(
			&task.DeletionID,
			&task.ObjectPath,
			&task.StorageNodeID,
			&task.Attempts,
			&task.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		task.Status = domain.StatusPending
		tasks = append(tasks, &task)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *GCRepository) CompleteGCTask(ctx context.Context, deletionID int64) error {
	query := `
		DELETE FROM gc_queue
		WHERE deletion_id = $1
	`

	db := GetDB(ctx, r.pool)

	tag, err := db.Exec(ctx, query, deletionID)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return errors.New("error: deletion not exists or already done")
	}

	return nil
}

func (r *GCRepository) IncrementGCTaskAttempts(ctx context.Context, deletionID int64) error {
	query := `
		UPDATE gc_queue
		SET attempts = attempts + 1
		WHERE deletion_id = $1
	`

	db := GetDB(ctx, r.pool)

	tag, err := db.Exec(ctx, query, deletionID)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return errors.New("error: deletion not exists")
	}

	return nil
}
