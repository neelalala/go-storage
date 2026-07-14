package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"

	"github.com/google/uuid"
	"github.com/neelalala/go-storage/internal/metadata/domain"
	metadatapb "github.com/neelalala/go-storage/pkg/proto/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	DefaultObjectsLimit  = 100
	DefaultObjectsOffset = 0
)

type MetadataService interface {
	InitUpload(ctx context.Context, bucket, key string, size uint64) (uuid.UUID, domain.Storage, error)
	CommitUpload(ctx context.Context, uploadID uuid.UUID, checksum uint32) error
	AbortUpload(ctx context.Context, uploadID uuid.UUID) error
	GetObject(ctx context.Context, bucket, key string) (domain.Object, domain.Storage, error)
	GetObjects(ctx context.Context, bucket, path string, limit, offset int) ([]domain.Object, error)
	DeleteObject(ctx context.Context, bucket, key string) error
}

type Server struct {
	metadatapb.UnimplementedMetadataServer
	addr       string
	grpcServer *grpc.Server
	service    MetadataService

	log *slog.Logger
}

func NewServer(addr string, service MetadataService, log *slog.Logger) *Server {
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

func (s *Server) InitUpload(ctx context.Context, req *metadatapb.InitUploadRequest) (*metadatapb.InitUploadResponse, error) {
	s.log.Debug("init upload request")

	bucket, key, size := req.GetBucket(), req.GetKey(), req.GetSize()

	id, node, err := s.service.InitUpload(ctx, bucket, key, size)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error initing upload: %v", err)
	}

	return &metadatapb.InitUploadResponse{
		UploadId: id.String(),
		StorageNode: &metadatapb.Node{
			Id:      node.ID.String(),
			Address: node.Address,
		},
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
		},
		StorageNode: &metadatapb.Node{
			Id:      node.ID.String(),
			Address: node.Address,
		},
	}, nil
}

func (s *Server) GetObjects(ctx context.Context, req *metadatapb.GetObjectsRequest) (*metadatapb.GetObjectsResponse, error) {
	s.log.Debug("get objects request")

	bucket, path := req.GetBucket(), req.GetPath()

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

	objs, err := s.service.GetObjects(ctx, bucket, path, limit, offset)
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

	return &metadatapb.GetObjectsResponse{
		Objects: pbobjects,
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
