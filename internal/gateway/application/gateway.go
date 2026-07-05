package application

import (
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

func (g *Gateway) PutObject(bucket, key string, data []byte) error {
	hash := string(g.hasher.Hash([]byte(bucket + key)))

	g.log.Debug("put object",
		"bucket", bucket,
		"key", key,
		"hash", hash,
		"data_size", len(data),
	)

	return g.storage.SaveObject(hash, data)
}

func (g *Gateway) GetObject(bucket, key string) ([]byte, error) {
	hash := string(g.hasher.Hash([]byte(bucket + key)))

	g.log.Debug("get object",
		"bucket", bucket,
		"key", key,
		"hash", hash,
	)

	return g.storage.GetObject(hash)
}

func (g *Gateway) DeleteObject(bucket, key string) error {
	hash := string(g.hasher.Hash([]byte(bucket + key)))

	g.log.Debug("delete object",
		"bucket", bucket,
		"key", key,
		"hash", hash,
	)

	return g.storage.DeleteObject(hash)
}
