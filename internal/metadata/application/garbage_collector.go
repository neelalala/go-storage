package application

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/neelalala/go-storage/internal/metadata/domain"
)

type StorageDeleter interface {
	DeleteObjectOn(ctx context.Context, nodeID uuid.UUID, path string) error
}

type GarbageCollector struct {
	gcRepo  domain.GCRepository
	storage StorageDeleter

	log *slog.Logger
}

func NewGarbageCollector(gcRepo domain.GCRepository, storage StorageDeleter, log *slog.Logger) *GarbageCollector {
	return &GarbageCollector{
		gcRepo:  gcRepo,
		storage: storage,
		log:     log,
	}
}

func (gc *GarbageCollector) Start(ctx context.Context, interval time.Duration, taskLimit int, taskTimeout time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			gc.log.Info("garbage collector stop")
			return
		case <-ticker.C:
			if err := gc.deleteObjects(ctx, taskLimit, taskTimeout); err != nil {
				if errors.Is(err, context.Canceled) {
					gc.log.Info("garbage collector canceled")
					return
				}
				gc.log.Error(
					"garbage collector",
					"method", "start",
					"context", "deleteObjects",
					"error", err,
				)
			}
		}
	}
}

func (gc *GarbageCollector) deleteObjects(ctx context.Context, limit int, timeout time.Duration) error {
	opCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	tasks, err := gc.gcRepo.GetPendingGCTasks(opCtx, limit)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			gc.log.Error(
				"garbage collector",
				"method", "delete objects",
				"context", "GarbageCollector.GetPendingGCTasks",
				"message", "deadline exceeded",
				"error", err,
			)
			return nil
		}
		return err
	}

	gc.log.Debug(
		"garbage collector",
		"method", "delete objects",
		"message", "got tasks",
		"count", len(tasks),
	)

	for _, task := range tasks {
		if opCtx.Err() != nil {
			return opCtx.Err()
		}

		err := gc.storage.DeleteObjectOn(opCtx, task.StorageNodeID, task.ObjectPath)
		if err != nil {
			if err := gc.gcRepo.IncrementGCTaskAttempts(opCtx, task.DeletionID); err != nil {
				gc.log.Error(
					"garbage collector",
					"method", "delete objects",
					"context", "GarbageCollector.IncrementGCTaskAttempts",
					"error", err,
				)
			}
			gc.log.Error(
				"garbage collector",
				"method", "delete objects",
				"context", "StorageDeleter.DeleteObject",
				"error", err,
			)
			continue
		}

		if err := gc.gcRepo.CompleteGCTask(opCtx, task.DeletionID); err != nil {
			gc.log.Error(
				"garbage collector",
				"method", "delete objects",
				"context", "GCRepository.CompleteGCTask",
				"error", err,
			)
		}
	}

	return nil
}
