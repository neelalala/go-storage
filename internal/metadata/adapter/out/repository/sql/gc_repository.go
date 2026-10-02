package sql

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neelalala/go-storage/internal/metadata/domain"
)

var _ domain.GCRepository = (*GCRepository)(nil)

type GCRepository struct {
	pool *pgxpool.Pool
}

func NewGCRepository(pool *pgxpool.Pool) *GCRepository {
	return &GCRepository{
		pool: pool,
	}
}

func (r *GCRepository) GetPendingGCTasks(ctx context.Context, limit int) ([]domain.GCTask, error) {
	if limit <= 0 {
		return nil, nil
	}

	query := `
		WITH next_tasks AS (
			SELECT deletion_id
			FROM gc_queue
			WHERE status = $1 OR (status = $2 AND updated_at < CURRENT_TIMESTAMP - INTERVAL '5 minutes')
			ORDER BY created_at, attempts 
			LIMIT $3 
			FOR UPDATE SKIP LOCKED
		)
		UPDATE gc_queue
		SET status = $2,
		    updated_at = CURRENT_TIMESTAMP
		FROM next_tasks
		WHERE gc_queue.deletion_id = next_tasks.deletion_id
		RETURNING gc_queue.deletion_id, gc_queue.object_path, gc_queue.storage_node_id, gc_queue.status, gc_queue.attempts, gc_queue.created_at, gc_queue.updated_at
	`

	db := GetDB(ctx, r.pool)

	rows, err := db.Query(ctx, query, domain.StatusPending, domain.StatusProcessing, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]domain.GCTask, 0, limit)

	for rows.Next() {
		var task domain.GCTask

		err := rows.Scan(
			&task.DeletionID,
			&task.ObjectPath,
			&task.StorageNodeID,
			&task.Status,
			&task.Attempts,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
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
		SET attempts = attempts + 1,
		    status = CASE WHEN attempts + 1 >= $2 THEN $3 ELSE $4 END,
		    updated_at = CURRENT_TIMESTAMP
		WHERE deletion_id = $1
	`

	db := GetDB(ctx, r.pool)

	tag, err := db.Exec(ctx, query, deletionID, domain.MaxGCAttempts, domain.StatusError, domain.StatusPending)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return errors.New("error: deletion not exists")
	}

	return nil
}
