package application

import (
	"context"
	"log/slog"

	"github.com/neelalala/go-storage/internal/storage/domain"
)

type Storage struct {
	store domain.Store

	log *slog.Logger
}

func NewStorage(store domain.Store, log *slog.Logger) *Storage {
	return &Storage{
		store: store,
		log:   log,
	}
}

func (s *Storage) SaveObject(ctx context.Context, obj domain.Object) error {
	s.log.Debug("save object",
		"name", obj.Name,
		"data_size", len(obj.Data),
	)

	return s.store.Save(ctx, obj)
}

func (s *Storage) GetObject(ctx context.Context, name string) (domain.Object, error) {
	s.log.Debug("get object",
		"name", name,
	)

	return s.store.Get(ctx, name)
}

func (s *Storage) DeleteObject(ctx context.Context, name string) error {
	s.log.Debug("delete object",
		"name", name,
	)

	return s.store.Delete(ctx, name)
}
