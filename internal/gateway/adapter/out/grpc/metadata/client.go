package metadata

import (
	"context"

	"github.com/google/uuid"
	"github.com/neelalala/go-storage/internal/gateway/domain"
	metadatapb "github.com/neelalala/go-storage/pkg/proto/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ domain.MetadataService = (*Client)(nil)

type Client struct {
	client metadatapb.MetadataClient
	conn   *grpc.ClientConn
}

func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		client: metadatapb.NewMetadataClient(conn),
		conn:   conn,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) InitUpload(ctx context.Context, bucket, key string, size uint64) (domain.Upload, domain.StorageNode, error) {
	req := &metadatapb.InitUploadRequest{
		Bucket: bucket,
		Key:    key,
		Size:   size,
	}

	resp, err := c.client.InitUpload(ctx, req)
	if err != nil {
		return domain.Upload{}, domain.StorageNode{}, err
	}

	nodeID, err := uuid.Parse(resp.StorageNode.GetId())
	if err != nil {
		return domain.Upload{}, domain.StorageNode{}, err
	}

	uploadID, err := uuid.Parse(resp.GetUploadId())
	if err != nil {
		return domain.Upload{}, domain.StorageNode{}, err
	}

	upload := domain.Upload{
		UploadID:      uploadID,
		Bucket:        bucket,
		Key:           key,
		ObjectPath:    resp.GetObjectPath(),
		Size:          size,
		StorageNodeID: nodeID,
	}

	storage := domain.StorageNode{
		ID:      nodeID,
		Address: resp.GetStorageNode().GetAddress(),
	}

	return upload, storage, nil
}

func (c *Client) CommitUpload(ctx context.Context, uploadID uuid.UUID, checksum uint32) error {
	req := &metadatapb.CommitUploadRequest{
		UploadId: uploadID.String(),
		Checksum: checksum,
	}

	_, err := c.client.CommitUpload(ctx, req)
	return err
}

func (c *Client) AbortUpload(ctx context.Context, uploadID uuid.UUID) error {
	req := &metadatapb.AbortUploadRequest{
		UploadId: uploadID.String(),
	}

	_, err := c.client.AbortUpload(ctx, req)
	return err
}

func (c *Client) GetObject(ctx context.Context, bucket, key string) (domain.ObjectMetadata, domain.StorageNode, error) {
	req := &metadatapb.GetObjectRequest{
		Bucket: bucket,
		Key:    key,
	}

	resp, err := c.client.GetObject(ctx, req)
	if err != nil {
		return domain.ObjectMetadata{}, domain.StorageNode{}, err
	}

	nodeID, err := uuid.Parse(resp.GetStorageNode().GetId())
	if err != nil {
		return domain.ObjectMetadata{}, domain.StorageNode{}, err
	}

	meta := domain.ObjectMetadata{
		Bucket:        bucket,
		Key:           key,
		ObjectPath:    resp.GetMetadata().GetObjectPath(),
		Size:          resp.GetMetadata().GetSize(),
		Checksum:      resp.GetMetadata().GetChecksum(),
		StorageNodeID: nodeID,
		CreatedAt:     resp.GetMetadata().GetCreatedAt().AsTime(),
		UpdatedAt:     resp.GetMetadata().GetUpdatedAt().AsTime(),
	}

	node := domain.StorageNode{
		ID:      nodeID,
		Address: resp.GetStorageNode().GetAddress(),
	}

	return meta, node, nil
}

func (c *Client) ListObjects(ctx context.Context, bucket, prefix, delimiter string, limit, offset int) ([]domain.ObjectMetadata, error) {
	req := &metadatapb.ListObjectsRequest{
		Bucket:    bucket,
		Prefix:    prefix,
		Delimiter: delimiter,
		Limit:     int32(limit),
		Offset:    int32(offset),
	}

	resp, err := c.client.ListObjects(ctx, req)
	if err != nil {
		return nil, err
	}

	objspb := resp.GetObjects()

	objs := make([]domain.ObjectMetadata, 0, len(objspb))

	for _, objpb := range objspb {
		nodeID, err := uuid.Parse(objpb.GetStorageNodeId())
		if err != nil {
			return nil, err
		}

		obj := domain.ObjectMetadata{
			Bucket:        objpb.GetBucket(),
			Key:           objpb.GetKey(),
			ObjectPath:    objpb.GetObjectPath(),
			Size:          objpb.GetSize(),
			Checksum:      objpb.GetChecksum(),
			StorageNodeID: nodeID,
			CreatedAt:     objpb.GetCreatedAt().AsTime(),
			UpdatedAt:     objpb.GetUpdatedAt().AsTime(),
		}

		objs = append(objs, obj)
	}

	return objs, nil
}

func (c *Client) DeleteObject(ctx context.Context, bucket, key string) error {
	req := &metadatapb.DeleteObjectRequest{
		Bucket: bucket,
		Key:    key,
	}

	_, err := c.client.DeleteObject(ctx, req)
	return err
}
