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
	return g.users.CreateUser(ctx, name)
}

func (g *Gateway) ListBuckets(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.BucketMetadata, error) {
	return g.metadata.ListBuckets(ctx, userID, limit, offset)
}

func (g *Gateway) CreateBucket(ctx context.Context, userID uuid.UUID, name string) (domain.BucketMetadata, error) {
	return g.metadata.CreateBucket(ctx, userID, name)
}

func (g *Gateway) ListObjects(
	ctx context.Context,
	userID uuid.UUID,
	bucket, prefix, delimiter string,
	limit, offset int,
) ([]domain.ObjectMetadata, []string, error) {
	return g.metadata.ListObjects(ctx, userID, bucket, prefix, delimiter, limit, offset)
}

func (g *Gateway) DeleteBucket(ctx context.Context, userID uuid.UUID, name string) error {
	return g.metadata.DeleteBucket(ctx, userID, name)
}

func (g *Gateway) PutObject(
	ctx context.Context,
	userID uuid.UUID,
	bucket string,
	key string,
	data []byte,
	contentType string,
	systemMetadata map[string]string,
	userMetadata map[string]string,
) error {
	upload, node, err := g.metadata.InitUpload(ctx, userID, bucket, key, uint64(len(data)), contentType, systemMetadata, userMetadata)
	if err != nil {
		return err
	}

	obj := domain.Object{
		Name: upload.ObjectPath,
		Data: data,
	}

	storage, err := g.nodes.GetStorage(node.Address)
	if err != nil {
		if err := g.metadata.AbortUpload(ctx, userID, upload.UploadID); err != nil {
			g.log.Error("Failed to abort upload", "error", err, "upload_id", upload.UploadID)
		}
		return err
	}

	hash, err := storage.SaveObject(ctx, obj)
	if err != nil {
		if err := g.metadata.AbortUpload(ctx, userID, upload.UploadID); err != nil {
			g.log.Error("Failed to abort upload", "error", err, "upload_id", upload.UploadID)
		}
		return err
	}

	return g.metadata.CommitUpload(ctx, userID, upload.UploadID, hash)
}

func (g *Gateway) GetObject(ctx context.Context, userID uuid.UUID, bucket, key string) (domain.ObjectMetadata, []byte, error) {
	meta, node, err := g.metadata.GetObject(ctx, userID, bucket, key)
	if err != nil {
		return domain.ObjectMetadata{}, nil, err
	}

	storage, err := g.nodes.GetStorage(node.Address)
	if err != nil {
		return domain.ObjectMetadata{}, nil, err
	}

	obj, err := storage.GetObject(ctx, meta.ObjectPath)
	if err != nil {
		return domain.ObjectMetadata{}, nil, err
	}

	return meta, obj.Data, nil
}

func (g *Gateway) DeleteObject(ctx context.Context, userID uuid.UUID, bucket, key string) error {
	err := g.metadata.DeleteObject(ctx, userID, bucket, key)
	if errors.Is(err, domain.ErrKeyNotExists) {
		return nil // idempotent method
	}
	return err
}
