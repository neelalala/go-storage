package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/neelalala/go-storage/internal/metadata/application"
	"github.com/neelalala/go-storage/internal/metadata/domain"
	metadatapb "github.com/neelalala/go-storage/pkg/proto/metadata"
)

// TODO: logging
type Server struct {
	metadatapb.UnimplementedMetadataServer
	addr       string
	grpcServer *grpc.Server
	service    *application.MetadataService

	log *slog.Logger
}

func NewServer(addr string, service *application.MetadataService, log *slog.Logger) *Server {
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
	limit, offset := int(req.GetLimit()), int(req.GetOffset())
	ownerID, err := uuid.Parse(req.GetOwnerId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error parsing owner id as uuid: %v", err)
	}

	buckets, err := s.service.ListBuckets(ctx, ownerID, limit, offset)
	if err != nil {
		if errors.Is(err, domain.ErrAccessDenied) {
			return nil, status.Errorf(codes.PermissionDenied, "error getting buckets: %v", err)
		}
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
	name := req.GetName()
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error parsing user id as uuid: %v", err)
	}

	bucket, err := s.service.CreateBucket(ctx, userID, name)
	if err != nil {
		if errors.Is(err, domain.ErrBucketAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, "error creating bucket: %v", err)
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

func (s *Server) DeleteBucket(ctx context.Context, req *metadatapb.DeleteBucketRequest) (*emptypb.Empty, error) {
	name := req.GetName()
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error parsing user id as uuid: %v", err)
	}

	err = s.service.DeleteBucket(ctx, userID, name)
	if err != nil {
		if errors.Is(err, domain.ErrAccessDenied) {
			return nil, status.Errorf(codes.PermissionDenied, "error deleting bucket: %v", err)
		}
		if errors.Is(err, domain.ErrBucketNotEmpty) {
			return nil, status.Errorf(codes.FailedPrecondition, "error deleting bucket: %v", err)
		}
		if errors.Is(err, domain.ErrBucketNotExists) {
			return nil, status.Errorf(codes.NotFound, "error deleting bucket: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "error deleting bucket: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) InitUpload(ctx context.Context, req *metadatapb.InitUploadRequest) (*metadatapb.InitUploadResponse, error) {
	bucket, key, size := req.GetBucket(), req.GetKey(), req.GetSize()
	contentType := req.GetContentType()
	systemMetadata, userMetadata := req.GetSystemMetadata(), req.GetUserMetadata()
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error parsing user id as uuid: %v", err)
	}

	upload, node, err := s.service.InitUpload(ctx, userID, bucket, key, size, contentType, systemMetadata, userMetadata)
	if err != nil {
		if errors.Is(err, domain.ErrAccessDenied) {
			return nil, status.Errorf(codes.PermissionDenied, "error initializing upload: %v", err)
		}
		if errors.Is(err, domain.ErrBucketNotExists) {
			return nil, status.Errorf(codes.NotFound, "error initializing upload: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "error initializing upload: %v", err)
	}

	return &metadatapb.InitUploadResponse{
		UploadId: upload.ID.String(),
		StorageNode: &metadatapb.Node{
			Id:      node.ID.String(),
			Address: node.Address,
		},
		ObjectPath: upload.ObjectPath,
	}, nil
}

func (s *Server) CommitUpload(ctx context.Context, req *metadatapb.CommitUploadRequest) (*emptypb.Empty, error) {
	uploadID, err := uuid.Parse(req.GetUploadId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error parsing upload id as uuid: %v", err)
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error parsing user id as uuid: %v", err)
	}
	hash := req.GetHash()

	err = s.service.CommitUpload(ctx, userID, uploadID, hash)
	if err != nil {
		if errors.Is(err, domain.ErrAccessDenied) {
			return nil, status.Errorf(codes.PermissionDenied, "error committing upload: %v", err)
		}
		if errors.Is(err, domain.ErrUploadNotExists) {
			return nil, status.Errorf(codes.NotFound, "error commiting upload: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "error commiting upload: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) AbortUpload(ctx context.Context, req *metadatapb.AbortUploadRequest) (*emptypb.Empty, error) {
	uploadID, err := uuid.Parse(req.GetUploadId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error parsing upload id as uuid: %v", err)
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error parsing user id as uuid: %v", err)
	}

	err = s.service.AbortUpload(ctx, userID, uploadID)
	if err != nil {
		if errors.Is(err, domain.ErrAccessDenied) {
			return nil, status.Errorf(codes.PermissionDenied, "error aborting upload: %v", err)
		}
		if errors.Is(err, domain.ErrUploadNotExists) {
			return nil, status.Errorf(codes.NotFound, "error aborting upload: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "error aborting upload: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetObject(ctx context.Context, req *metadatapb.GetObjectRequest) (*metadatapb.GetObjectResponse, error) {
	bucket, key := req.GetBucket(), req.GetKey()
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error parsing user id as uuid: %v", err)
	}

	obj, node, err := s.service.GetObject(ctx, userID, bucket, key)
	if err != nil {
		if errors.Is(err, domain.ErrAccessDenied) {
			return nil, status.Errorf(codes.PermissionDenied, "error getting object: %v", err)
		}
		if errors.Is(err, domain.ErrBucketNotExists) {
			return nil, status.Errorf(codes.FailedPrecondition, "error getting object: %v", err)
		}
		if errors.Is(err, domain.ErrObjectNotFound) {
			return nil, status.Errorf(codes.NotFound, "error getting object: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "error getting object: %v", err)
	}

	return &metadatapb.GetObjectResponse{
		Metadata: &metadatapb.ObjectMetadata{
			Bucket:         obj.Bucket,
			Key:            obj.Key,
			Size:           obj.Size,
			CreatedAt:      timestamppb.New(obj.CreatedAt),
			UpdatedAt:      timestamppb.New(obj.UpdatedAt),
			StorageNodeId:  obj.StorageNodeID.String(),
			ObjectPath:     obj.ObjectPath,
			ContentType:    obj.ContentType,
			Hash:           obj.Hash,
			SystemMetadata: obj.SystemMetadata,
			UserMetadata:   obj.UserMetadata,
			OwnerId:        obj.OwnerID.String(),
		},
		StorageNode: &metadatapb.Node{
			Id:      node.ID.String(),
			Address: node.Address,
		},
	}, nil
}

func (s *Server) ListObjects(ctx context.Context, req *metadatapb.ListObjectsRequest) (*metadatapb.ListObjectsResponse, error) {
	bucket, prefix, delimiter := req.GetBucket(), req.GetPrefix(), req.GetDelimiter()
	limit, offset := int(req.GetLimit()), int(req.GetOffset())
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error parsing user id as uuid: %v", err)
	}

	objs, prefixes, err := s.service.GetObjects(ctx, userID, bucket, prefix, delimiter, limit, offset)
	if err != nil {
		if errors.Is(err, domain.ErrAccessDenied) {
			return nil, status.Errorf(codes.PermissionDenied, "error getting objects: %v", err)
		}
		if errors.Is(err, domain.ErrBucketNotExists) {
			return nil, status.Errorf(codes.NotFound, "error getting objects: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "error getting objects: %v", err)
	}

	pbobjects := make([]*metadatapb.ObjectMetadata, 0, len(objs))
	for _, obj := range objs {
		pbobject := &metadatapb.ObjectMetadata{
			Bucket:         obj.Bucket,
			Key:            obj.Key,
			Size:           obj.Size,
			CreatedAt:      timestamppb.New(obj.CreatedAt),
			UpdatedAt:      timestamppb.New(obj.UpdatedAt),
			StorageNodeId:  obj.StorageNodeID.String(),
			ObjectPath:     obj.ObjectPath,
			ContentType:    obj.ContentType,
			Hash:           obj.Hash,
			SystemMetadata: obj.SystemMetadata,
			UserMetadata:   obj.UserMetadata,
			OwnerId:        obj.OwnerID.String(),
		}
		pbobjects = append(pbobjects, pbobject)
	}

	return &metadatapb.ListObjectsResponse{
		Objects:        pbobjects,
		CommonPrefixes: prefixes,
	}, nil
}

func (s *Server) DeleteObject(ctx context.Context, req *metadatapb.DeleteObjectRequest) (*emptypb.Empty, error) {
	bucket, key := req.GetBucket(), req.GetKey()
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error parsing user id as uuid: %v", err)
	}

	if err := s.service.DeleteObject(ctx, userID, bucket, key); err != nil {
		if errors.Is(err, domain.ErrAccessDenied) {
			return nil, status.Errorf(codes.PermissionDenied, "error deleting object: %v", err)
		}
		if errors.Is(err, domain.ErrBucketNotExists) {
			return nil, status.Errorf(codes.FailedPrecondition, "error deleting object: %v", err)
		}
		if errors.Is(err, domain.ErrObjectNotFound) {
			return nil, status.Errorf(codes.NotFound, "error deleting object: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "error deleting object: %v", err)
	}

	return &emptypb.Empty{}, nil
}
