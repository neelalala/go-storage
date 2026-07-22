package application

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"

	"github.com/neelalala/go-storage/internal/gateway/domain"
)

type Gateway struct {
	metadata domain.MetadataService
	users    domain.UserService
	nodes    domain.StorageNodeManager

	log *slog.Logger
}

func NewGateway(metadata domain.MetadataService, users domain.UserService, nodes domain.StorageNodeManager, log *slog.Logger) *Gateway {
	return &Gateway{
		metadata: metadata,
		users:    users,
		nodes:    nodes,
		log:      log,
	}
}

func (g *Gateway) CreateUser(ctx context.Context, name string) (domain.User, error) {
	g.log.Debug("create user", "name", name)

	user, err := g.users.CreateUser(ctx, name)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return domain.User{}, domain.ErrUserAlreadyExists
		}
	}

	return user, nil
}

func (g *Gateway) GetUserByName(ctx context.Context, name string) (domain.User, error) {
	g.log.Debug("get user", "name", name)

	return g.users.GetUserByName(ctx, name)
}

func (g *Gateway) ListBuckets(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.BucketMetadata, error) {
	g.log.Debug("list buckets", "userID", userID, "limit", limit, "offset", offset)

	return g.metadata.ListBuckets(ctx, userID, limit, offset)
}

func (g *Gateway) PutObject(ctx context.Context, userID uuid.UUID, bucket, key string, data []byte, contentType string) error {
	g.log.Debug("put object",
		"bucket", bucket,
		"key", key,
		"data_size", len(data),
	)

	upload, node, err := g.metadata.InitUpload(ctx, bucket, key, uint64(len(data)))
	if err != nil {
		return err
	}

	g.log.Debug("put object", "object_path", upload.ObjectPath)

	obj := domain.Object{
		Name: upload.ObjectPath,
		Data: data,
	}

	storage, err := g.nodes.GetStorage(node.Address)
	if err != nil {
		g.metadata.AbortUpload(ctx, upload.UploadID)
		return err
	}

	checksum, err := storage.SaveObject(ctx, obj)
	if err != nil {
		g.metadata.AbortUpload(ctx, upload.UploadID)
		return err
	}

	return g.metadata.CommitUpload(ctx, upload.UploadID, checksum)
}

func (g *Gateway) GetObject(ctx context.Context, bucket, key string) ([]byte, error) {
	g.log.Debug("get object",
		"bucket", bucket,
		"key", key,
	)

	meta, node, err := g.metadata.GetObject(ctx, bucket, key)
	if err != nil {
		return nil, err
	}

	storage, err := g.nodes.GetStorage(node.Address)
	if err != nil {
		return nil, err
	}

	obj, err := storage.GetObject(ctx, meta.ObjectPath)
	if err != nil {
		return nil, err
	}

	return obj.Data, nil
}

func (g *Gateway) DeleteObject(ctx context.Context, bucket, key string) error {
	g.log.Debug("delete object",
		"bucket", bucket,
		"key", key,
	)

	return g.metadata.DeleteObject(ctx, bucket, key)
}
