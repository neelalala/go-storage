package application

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/neelalala/go-storage/internal/metadata/domain"
)

type StorageDeleter interface {
	DeleteObject(ctx context.Context, path string) error
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
			if err := gc.deleteObjects(taskLimit, taskTimeout); err != nil {
				gc.log.Error("garbage collector",
					"method", "start",
					"context", "deleteObjects",
					"error", err,
				)
			}
		}
	}
}

func (gc *GarbageCollector) deleteObjects(limit int, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tasks, err := gc.gcRepo.GetPendingGCTasks(ctx, limit)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			gc.log.Error("garbage collector",
				"method", "delete objects",
				"context", "GarbageCollector.GetPendingGCTasks",
				"message", "deadline exceeded",
				"error", err,
			)
			return nil
		}
		return err
	}

	gc.log.Debug("garbage collector",
		"method", "delete objects",
		"message", "got tasks",
		"count", len(tasks),
	)

	for _, task := range tasks {
		err := gc.storage.DeleteObject(ctx, task.ObjectPath)
		if err != nil {
			if err := gc.gcRepo.IncrementGCTaskAttempts(ctx, task.DeletionID); err != nil {
				gc.log.Error("garbage collector",
					"method", "delete objects",
					"context", "GarbageCollector.IncrementGCTaskAttempts",
					"error", err,
				)
			}
			gc.log.Error("garbage collector",
				"method", "delete objects",
				"context", "StorageDeleter.DeleteObject",
				"error", err,
			)
		}

		if err := gc.gcRepo.CompleteGCTask(ctx, task.DeletionID); err != nil {
			gc.log.Error("garbage collector",
				"method", "delete objects",
				"context", "GCRepository.CompleteGCTask",
				"error", err,
			)
		}
	}

	return nil
}
