package metadata

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/neelalala/go-storage/internal/gateway/domain"
	metadatapb "github.com/neelalala/go-storage/pkg/proto/metadata"
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

func (c *Client) ListBuckets(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.BucketMetadata, error) {
	req := &metadatapb.ListBucketsRequest{
		Limit:   int32(limit),
		Offset:  int32(offset),
		OwnerId: userID.String(),
	}

	resp, err := c.client.ListBuckets(ctx, req)
	if err != nil {
		if status.Code(err) == codes.PermissionDenied { // why can user get this error?
			return nil, fmt.Errorf("%w: %v", domain.ErrAccessDenied, err)
		}
		return nil, err
	}

	pbbuckets := resp.GetBuckets()

	buckets := make([]domain.BucketMetadata, 0, len(pbbuckets))
	for _, pbbucket := range pbbuckets {
		bucket := domain.BucketMetadata{
			Name:      pbbucket.GetName(),
			CreatedAt: pbbucket.CreatedAt.AsTime(),
		}

		buckets = append(buckets, bucket)
	}

	return buckets, nil
}

func (c *Client) CreateBucket(ctx context.Context, userID uuid.UUID, name string) (domain.BucketMetadata, error) {
	req := &metadatapb.CreateBucketRequest{
		Name:   name,
		UserId: userID.String(),
	}

	resp, err := c.client.CreateBucket(ctx, req)
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return domain.BucketMetadata{}, fmt.Errorf("%w: %v", domain.ErrBucketAlreadyExists, err)
		}
		return domain.BucketMetadata{}, err
	}

	bucket := domain.BucketMetadata{
		Name:      resp.GetBucket().GetName(),
		CreatedAt: resp.GetBucket().GetCreatedAt().AsTime(),
	}

	return bucket, nil
}

func (c *Client) HeadBucket(ctx context.Context, userID uuid.UUID, bucket string) (domain.BucketMetadata, error) {
	req := &metadatapb.HeadBucketRequest{
		Bucket: bucket,
		UserId: userID.String(),
	}

	resp, err := c.client.HeadBucket(ctx, req)
	if err != nil {
		if status.Code(err) == codes.PermissionDenied {
			return domain.BucketMetadata{}, fmt.Errorf("%w: %v", domain.ErrAccessDenied, err)
		}
		if status.Code(err) == codes.NotFound {
			return domain.BucketMetadata{}, fmt.Errorf("%w: %v", domain.ErrBucketNotExists, err)
		}
		return domain.BucketMetadata{}, err
	}

	ownerID, err := uuid.Parse(resp.GetMetadata().GetOwnerId())
	if err != nil {
		return domain.BucketMetadata{}, err
	}

	meta := domain.BucketMetadata{
		Name:      resp.GetMetadata().Name,
		OwnerID:   ownerID,
		CreatedAt: resp.GetMetadata().CreatedAt.AsTime(),
	}

	return meta, nil
}

func (c *Client) DeleteBucket(ctx context.Context, userID uuid.UUID, name string) error {
	req := &metadatapb.DeleteBucketRequest{
		Name:   name,
		UserId: userID.String(),
	}

	_, err := c.client.DeleteBucket(ctx, req)
	if err != nil {
		if status.Code(err) == codes.PermissionDenied {
			return fmt.Errorf("%w: %v", domain.ErrAccessDenied, err)
		}
		if status.Code(err) == codes.FailedPrecondition {
			return fmt.Errorf("%w: %v", domain.ErrBucketNotEmpty, err)
		}
		if status.Code(err) == codes.NotFound {
			return fmt.Errorf("%w: %v", domain.ErrBucketNotExists, err)
		}
		return err
	}

	return nil
}

func (c *Client) ListObjects(ctx context.Context, userID uuid.UUID, bucket, prefix, delimiter string, limit, offset int) ([]domain.ObjectMetadata, []string, error) {
	req := &metadatapb.ListObjectsRequest{
		Bucket:    bucket,
		Prefix:    prefix,
		Delimiter: delimiter,
		Limit:     int32(limit),
		Offset:    int32(offset),
		UserId:    userID.String(),
	}

	resp, err := c.client.ListObjects(ctx, req)
	if err != nil {
		if status.Code(err) == codes.PermissionDenied {
			return nil, nil, fmt.Errorf("%w: %v", domain.ErrAccessDenied, err)
		}
		if status.Code(err) == codes.NotFound {
			return nil, nil, fmt.Errorf("%w: %v", domain.ErrBucketNotExists, err)
		}
		return nil, nil, err
	}

	objspb := resp.GetObjects()
	prefixes := resp.GetCommonPrefixes()

	objs := make([]domain.ObjectMetadata, 0, len(objspb))

	for _, objpb := range objspb {
		nodeID, err := uuid.Parse(objpb.GetStorageNodeId())
		if err != nil {
			return nil, nil, err
		}

		ownerID, err := uuid.Parse(objpb.GetOwnerId())
		if err != nil {
			return nil, nil, err
		}

		obj := domain.ObjectMetadata{
			Bucket:         objpb.GetBucket(),
			Key:            objpb.GetKey(),
			ObjectPath:     objpb.GetObjectPath(),
			Size:           objpb.GetSize(),
			StorageNodeID:  nodeID,
			CreatedAt:      objpb.GetCreatedAt().AsTime(),
			UpdatedAt:      objpb.GetUpdatedAt().AsTime(),
			ContentType:    objpb.GetContentType(),
			Hash:           objpb.GetHash(),
			SystemMetadata: objpb.GetSystemMetadata(),
			UserMetadata:   objpb.GetUserMetadata(),
			OwnerID:        ownerID,
		}

		objs = append(objs, obj)
	}

	return objs, prefixes, nil
}

func (c *Client) InitUpload(
	ctx context.Context,
	userID uuid.UUID,
	bucket, key string,
	size uint64,
	contentType string,
	systemMetadata map[string]string,
	userMetadata map[string]string,
) (domain.Upload, domain.StorageNode, error) {
	req := &metadatapb.InitUploadRequest{
		Bucket:         bucket,
		Key:            key,
		Size:           size,
		ContentType:    contentType,
		SystemMetadata: systemMetadata,
		UserMetadata:   userMetadata,
		UserId:         userID.String(),
	}

	resp, err := c.client.InitUpload(ctx, req)
	if err != nil {
		if status.Code(err) == codes.PermissionDenied {
			return domain.Upload{}, domain.StorageNode{}, fmt.Errorf("%w: %v", domain.ErrAccessDenied, err)
		}
		if status.Code(err) == codes.NotFound {
			return domain.Upload{}, domain.StorageNode{}, fmt.Errorf("%w: %v", domain.ErrBucketNotExists, err)
		}
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
		UploadID:       uploadID,
		Bucket:         bucket,
		Key:            key,
		ObjectPath:     resp.GetObjectPath(),
		Size:           size,
		StorageNodeID:  nodeID,
		CreatedAt:      resp.GetCreatedAt().AsTime(),
		ContentType:    contentType,
		SystemMetadata: systemMetadata,
		UserMetadata:   userMetadata,
		OwnerID:        userID,
	}

	storage := domain.StorageNode{
		ID:      nodeID,
		Address: resp.GetStorageNode().GetAddress(),
	}

	return upload, storage, nil
}

func (c *Client) CommitUpload(ctx context.Context, userID, uploadID uuid.UUID, hash string) error {
	req := &metadatapb.CommitUploadRequest{
		UploadId: uploadID.String(),
		UserId:   userID.String(),
		Hash:     hash,
	}

	_, err := c.client.CommitUpload(ctx, req)
	if err != nil {
		if status.Code(err) == codes.PermissionDenied {
			return fmt.Errorf("%w: %v", domain.ErrAccessDenied, err)
		}
		if status.Code(err) == codes.NotFound {
			return fmt.Errorf("%w: %v", domain.ErrUploadNotExists, err)
		}
		return err
	}

	return nil
}

func (c *Client) AbortUpload(ctx context.Context, userID, uploadID uuid.UUID) error {
	req := &metadatapb.AbortUploadRequest{
		UploadId: uploadID.String(),
		UserId:   userID.String(),
	}

	_, err := c.client.AbortUpload(ctx, req)
	if err != nil {
		if status.Code(err) == codes.PermissionDenied {
			return fmt.Errorf("%w: %v", domain.ErrAccessDenied, err)
		}
		if status.Code(err) == codes.NotFound {
			return fmt.Errorf("%w: %v", domain.ErrUploadNotExists, err)
		}
		return err
	}

	return nil
}

func (c *Client) HeadObject(ctx context.Context, userID uuid.UUID, bucket, key string) (domain.ObjectMetadata, error) {
	req := &metadatapb.HeadObjectRequest{
		Bucket: bucket,
		Key:    key,
		UserId: userID.String(),
	}

	resp, err := c.client.HeadObject(ctx, req)
	if err != nil {
		if status.Code(err) == codes.PermissionDenied {
			return domain.ObjectMetadata{}, fmt.Errorf("%w: %v", domain.ErrAccessDenied, err)
		}
		if status.Code(err) == codes.FailedPrecondition {
			return domain.ObjectMetadata{}, fmt.Errorf("%w: %v", domain.ErrBucketNotExists, err)
		}
		if status.Code(err) == codes.NotFound {
			return domain.ObjectMetadata{}, fmt.Errorf("%w: %v", domain.ErrKeyNotExists, err)
		}
		return domain.ObjectMetadata{}, err
	}

	storageNodeID, err := uuid.Parse(resp.GetMetadata().GetStorageNodeId())
	if err != nil {
		return domain.ObjectMetadata{}, err
	}

	ownerID, err := uuid.Parse(resp.GetMetadata().GetOwnerId())
	if err != nil {
		return domain.ObjectMetadata{}, err
	}

	meta := domain.ObjectMetadata{
		Bucket:         resp.GetMetadata().GetBucket(),
		Key:            resp.GetMetadata().GetKey(),
		ObjectPath:     resp.GetMetadata().GetObjectPath(),
		Size:           resp.GetMetadata().GetSize(),
		StorageNodeID:  storageNodeID,
		CreatedAt:      resp.GetMetadata().GetCreatedAt().AsTime(),
		UpdatedAt:      resp.GetMetadata().GetUpdatedAt().AsTime(),
		ContentType:    resp.GetMetadata().GetContentType(),
		Hash:           resp.GetMetadata().GetHash(),
		SystemMetadata: resp.GetMetadata().GetSystemMetadata(),
		UserMetadata:   resp.GetMetadata().GetUserMetadata(),
		OwnerID:        ownerID,
	}

	return meta, nil
}

func (c *Client) GetObject(ctx context.Context, userID uuid.UUID, bucket, key string) (domain.ObjectMetadata, domain.StorageNode, error) {
	req := &metadatapb.GetObjectRequest{
		Bucket: bucket,
		Key:    key,
		UserId: userID.String(),
	}

	resp, err := c.client.GetObject(ctx, req)
	if err != nil {
		if status.Code(err) == codes.PermissionDenied {
			return domain.ObjectMetadata{}, domain.StorageNode{}, fmt.Errorf("%w: %v", domain.ErrAccessDenied, err)
		}
		if status.Code(err) == codes.FailedPrecondition {
			return domain.ObjectMetadata{}, domain.StorageNode{}, fmt.Errorf("%w: %v", domain.ErrBucketNotExists, err)
		}
		if status.Code(err) == codes.NotFound {
			return domain.ObjectMetadata{}, domain.StorageNode{}, fmt.Errorf("%w: %v", domain.ErrKeyNotExists, err)
		}
		return domain.ObjectMetadata{}, domain.StorageNode{}, err
	}

	nodeID, err := uuid.Parse(resp.GetStorageNode().GetId())
	if err != nil {
		return domain.ObjectMetadata{}, domain.StorageNode{}, err
	}

	ownerID, err := uuid.Parse(resp.GetMetadata().GetOwnerId())
	if err != nil {
		return domain.ObjectMetadata{}, domain.StorageNode{}, err
	}

	meta := domain.ObjectMetadata{
		Bucket:         bucket,
		Key:            key,
		ObjectPath:     resp.GetMetadata().GetObjectPath(),
		Size:           resp.GetMetadata().GetSize(),
		StorageNodeID:  nodeID,
		CreatedAt:      resp.GetMetadata().GetCreatedAt().AsTime(),
		UpdatedAt:      resp.GetMetadata().GetUpdatedAt().AsTime(),
		ContentType:    resp.GetMetadata().GetContentType(),
		Hash:           resp.GetMetadata().GetHash(),
		SystemMetadata: resp.GetMetadata().GetSystemMetadata(),
		UserMetadata:   resp.GetMetadata().GetUserMetadata(),
		OwnerID:        ownerID,
	}

	node := domain.StorageNode{
		ID:      nodeID,
		Address: resp.GetStorageNode().GetAddress(),
	}

	return meta, node, nil
}

func (c *Client) DeleteObject(ctx context.Context, userID uuid.UUID, bucket, key string) error {
	req := &metadatapb.DeleteObjectRequest{
		Bucket: bucket,
		Key:    key,
		UserId: userID.String(),
	}

	_, err := c.client.DeleteObject(ctx, req)
	if err != nil {
		if status.Code(err) == codes.PermissionDenied {
			return fmt.Errorf("%w: %v", domain.ErrAccessDenied, err)
		}
		if status.Code(err) == codes.FailedPrecondition {
			return fmt.Errorf("%w: %v", domain.ErrBucketNotExists, err)
		}
		if status.Code(err) == codes.NotFound {
			return fmt.Errorf("%w: %v", domain.ErrKeyNotExists, err)
		}
		return err
	}

	return nil
}
