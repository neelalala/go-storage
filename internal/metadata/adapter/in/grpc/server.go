package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"

	"github.com/google/uuid"
	"github.com/neelalala/go-storage/internal/metadata/application"
	"github.com/neelalala/go-storage/internal/metadata/domain"
	metadatapb "github.com/neelalala/go-storage/pkg/proto/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	DefaultBucketsLimit  = 100
	DefaultBucketsOffset = 0
	DefaultObjectsLimit  = 100
	DefaultObjectsOffset = 0
)

type Server struct {
	metadatapb.UnimplementedMetadataServer
	addr       string
	grpcServer *grpc.Server
	service    application.MetadataService

	log *slog.Logger
}

func NewServer(addr string, service application.MetadataService, log *slog.Logger) *Server {
	grpcServer := grpc.NewServer()

	server := &Server{
		addr:       addr,
		grpcServer: grpcServer,
		service:    service,
		log:        log,
	}

	metadatapb.RegisterMetadataServer(grpcServer, server)

	return server
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to lister on address %s: %w", s.addr, err)
	}

	s.log.Info("gRPC server is running", "address", s.addr)
	if err := s.grpcServer.Serve(listener); err != nil {
		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.log.Info("Shutting down gRPC server...")

	stopped := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		s.grpcServer.Stop()
		return ctx.Err()
	case <-stopped:
		return nil
	}
}

func (s *Server) ListBuckets(ctx context.Context, req *metadatapb.ListBucketsRequest) (*metadatapb.ListBucketsResponse, error) {
	s.log.Debug("list buckets request")

	var (
		limit  int
		offset int
	)

	if req.Limit != nil {
		limit = int(req.GetLimit())
	} else {
		limit = DefaultBucketsLimit
	}

	if req.Offset != nil {
		offset = int(req.GetOffset())
	} else {
		offset = DefaultBucketsOffset
	}

	buckets, err := s.service.ListBuckets(ctx, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error getting buckets: %v", err)
	}

	pbbuckets := make([]*metadatapb.BucketMetadata, 0, len(buckets))
	for _, bucket := range buckets {
		pbbucket := &metadatapb.BucketMetadata{
			Name:      bucket.Name,
			CreatedAt: timestamppb.New(bucket.CreatedAt),
		}
		pbbuckets = append(pbbuckets, pbbucket)
	}

	return &metadatapb.ListBucketsResponse{
		Buckets: pbbuckets,
	}, nil
}

func (s *Server) CreateBucket(ctx context.Context, req *metadatapb.CreateBucketRequest) (*metadatapb.CreateBucketResponse, error) {
	s.log.Debug("create bucket request")

	name := req.GetName()

	bucket, err := s.service.CreateBucket(ctx, name)
	if err != nil {
		if errors.Is(err, domain.ErrBucketExists) {
			return nil, status.Errorf(codes.InvalidArgument, "error creating bucket: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "error creating bucket: %v", err)
	}

	return &metadatapb.CreateBucketResponse{
		Bucket: &metadatapb.BucketMetadata{
			Name:      bucket.Name,
			CreatedAt: timestamppb.New(bucket.CreatedAt),
		},
	}, nil
}

func (s *Server) ListObjects(ctx context.Context, req *metadatapb.ListObjectsRequest) (*metadatapb.ListObjectsResponse, error) {
	s.log.Debug("get objects request")

	bucket, prefix, delimiter := req.GetBucket(), req.GetPrefix(), req.GetDelimiter()

	var (
		limit  int
		offset int
	)

	if req.Limit != nil {
		limit = int(req.GetLimit())
	} else {
		limit = DefaultObjectsLimit
	}

	if req.Offset != nil {
		offset = int(req.GetOffset())
	} else {
		offset = DefaultObjectsOffset
	}

	objs, err := s.service.GetObjects(ctx, bucket, prefix, delimiter, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error getting objects: %v", err)
	}

	pbobjects := make([]*metadatapb.ObjectMetadata, 0, len(objs))
	for _, obj := range objs {
		pbobject := &metadatapb.ObjectMetadata{
			Bucket:        obj.Bucket,
			Key:           obj.Key,
			Size:          obj.Size,
			Checksum:      obj.Checksum,
			CreatedAt:     timestamppb.New(obj.CreatedAt),
			UpdatedAt:     timestamppb.New(obj.UpdatedAt),
			StorageNodeId: obj.StorageNodeID.String(),
		}
		pbobjects = append(pbobjects, pbobject)
	}

	return &metadatapb.ListObjectsResponse{
		Objects: pbobjects,
	}, nil

}

func (s *Server) DeleteBucket(ctx context.Context, req *metadatapb.DeleteBucketRequest) (*emptypb.Empty, error) {
	s.log.Debug("delete bucket request")

	name := req.GetName()

	err := s.service.DeleteBucket(ctx, name)
	if err != nil {
		if errors.Is(err, domain.ErrBucketNotExists) {
			return nil, status.Errorf(codes.NotFound, "error deleting bucket: %v", err)
		}
		if errors.Is(err, domain.ErrBucketNotEmpty) {
			return nil, status.Errorf(codes.InvalidArgument, "error deleting bucket: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "error deleting bucket: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) HeadBucket(ctx context.Context, req *metadatapb.HeadBucketRequest) (*metadatapb.HeadBucketResponse, error) {
	s.log.Debug("head bucket request")

	name := req.GetName()

	bucket, err := s.service.GetBucket(ctx, name)
	if err != nil {
		if errors.Is(err, domain.ErrBucketNotExists) {
			return nil, status.Errorf(codes.NotFound, "error getting bucket head: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "error getting bucket head: %v", err)
	}

	return &metadatapb.HeadBucketResponse{
		BucketMeta: &metadatapb.BucketMetadata{
			Name:      bucket.Name,
			CreatedAt: timestamppb.New(bucket.CreatedAt),
		},
	}, nil
}

func (s *Server) InitUpload(ctx context.Context, req *metadatapb.InitUploadRequest) (*metadatapb.InitUploadResponse, error) {
	s.log.Debug("init upload request")

	bucket, key, size := req.GetBucket(), req.GetKey(), req.GetSize()

	upload, node, err := s.service.InitUpload(ctx, bucket, key, size)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error initing upload: %v", err)
	}

	return &metadatapb.InitUploadResponse{
		UploadId: upload.UploadID.String(),
		StorageNode: &metadatapb.Node{
			Id:      node.ID.String(),
			Address: node.Address,
		},
		ObjectPath: upload.ObjectPath,
	}, nil
}

func (s *Server) CommitUpload(ctx context.Context, req *metadatapb.CommitUploadRequest) (*emptypb.Empty, error) {
	s.log.Debug("commit upload request")

	id, checksum := req.GetUploadId(), req.GetChecksum()
	uuid, err := uuid.Parse(id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error commiting upload: invalid uploadID: %v", err)
	}

	err = s.service.CommitUpload(ctx, uuid, checksum)
	if err != nil {
		if errors.Is(err, domain.ErrUploadNotFound) {
			return nil, status.Errorf(codes.NotFound, "error commiting upload: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "error commiting upload: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) AbortUpload(ctx context.Context, req *metadatapb.AbortUploadRequest) (*emptypb.Empty, error) {
	s.log.Debug("abort upload request")

	id := req.GetUploadId()
	uuid, err := uuid.Parse(id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "error aborting upload: invalid uploadID: %v", err)
	}

	err = s.service.AbortUpload(ctx, uuid)
	if err != nil {
		if errors.Is(err, domain.ErrUploadNotFound) {
			return nil, status.Errorf(codes.NotFound, "error aborting upload: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "error aborting upload: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetObject(ctx context.Context, req *metadatapb.GetObjectRequest) (*metadatapb.GetObjectResponse, error) {
	s.log.Debug("get object request")

	bucket, key := req.GetBucket(), req.GetKey()

	obj, node, err := s.service.GetObject(ctx, bucket, key)
	if err != nil {
		if errors.Is(err, domain.ErrObjectNotFound) {
			return nil, status.Errorf(codes.NotFound, "error getting object: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "error getting object: %v", err)
	}

	return &metadatapb.GetObjectResponse{
		Metadata: &metadatapb.ObjectMetadata{
			Bucket:        obj.Bucket,
			Key:           obj.Key,
			Size:          obj.Size,
			Checksum:      obj.Checksum,
			CreatedAt:     timestamppb.New(obj.CreatedAt),
			UpdatedAt:     timestamppb.New(obj.UpdatedAt),
			StorageNodeId: obj.StorageNodeID.String(),
			ObjectPath:    obj.ObjectPath,
		},
		StorageNode: &metadatapb.Node{
			Id:      node.ID.String(),
			Address: node.Address,
		},
	}, nil
}

func (s *Server) DeleteObject(ctx context.Context, req *metadatapb.DeleteObjectRequest) (*emptypb.Empty, error) {
	s.log.Debug("delete object request")

	bucket, key := req.GetBucket(), req.GetKey()

	if err := s.service.DeleteObject(ctx, bucket, key); err != nil {
		if errors.Is(err, domain.ErrObjectNotFound) {
			return nil, status.Errorf(codes.NotFound, "error deleting object: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "error deleting object: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) HeadObject(ctx context.Context, req *metadatapb.HeadObjectRequest) (*metadatapb.HeadObjectResponse, error) {
	s.log.Debug("head object request")

	bucket, key := req.GetBucket(), req.GetKey()

	obj, err := s.service.HeadObject(ctx, bucket, key)
	if err != nil {
		if errors.Is(err, domain.ErrObjectNotFound) {
			return nil, status.Errorf(codes.NotFound, "error getting object head: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "error getting object head: %v", err)
	}

	return &metadatapb.HeadObjectResponse{
		ObjectMeta: &metadatapb.ObjectMetadata{
			Bucket:        obj.Bucket,
			Key:           obj.Key,
			Size:          obj.Size,
			Checksum:      obj.Checksum,
			CreatedAt:     timestamppb.New(obj.CreatedAt),
			UpdatedAt:     timestamppb.New(obj.UpdatedAt),
			StorageNodeId: obj.StorageNodeID.String(),
			ObjectPath:    obj.ObjectPath,
		},
	}, nil
}
