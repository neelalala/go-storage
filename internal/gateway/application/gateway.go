package application

import (
	"context"
	"log/slog"

	"github.com/neelalala/go-storage/internal/gateway/domain"
)

type Gateway struct {
	storage domain.Storage
	hasher  domain.Hasher

	log *slog.Logger
}

func NewGateway(storage domain.Storage, hasher domain.Hasher, log *slog.Logger) *Gateway {
	return &Gateway{
		storage: storage,
		hasher:  hasher,
		log:     log,
	}
}

func (g *Gateway) PutObject(ctx context.Context, bucket, key string, data []byte) error {
	hash := string(g.hasher.Hash([]byte(bucket + key)))

	g.log.Debug("put object",
		"bucket", bucket,
		"key", key,
		"hash", hash,
		"data_size", len(data),
	)

	obj := domain.Object{
		Name: hash,
		Data: data,
	}

	return g.storage.SaveObject(ctx, obj)
}

func (g *Gateway) GetObject(ctx context.Context, bucket, key string) ([]byte, error) {
	hash := string(g.hasher.Hash([]byte(bucket + key)))

	g.log.Debug("get object",
		"bucket", bucket,
		"key", key,
		"hash", hash,
	)

	obj, err := g.storage.GetObject(ctx, hash)
	if err != nil {
		return nil, err
	}

	return obj.Data, nil
}

func (g *Gateway) DeleteObject(ctx context.Context, bucket, key string) error {
	hash := string(g.hasher.Hash([]byte(bucket + key)))

	g.log.Debug("delete object",
		"bucket", bucket,
		"key", key,
		"hash", hash,
	)

	return g.storage.DeleteObject(ctx, hash)
}
