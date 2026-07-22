package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/neelalala/go-storage/internal/storage/domain"
	storagepb "github.com/neelalala/go-storage/pkg/proto/storage"
)

type Storage interface {
	SaveObject(ctx context.Context, obj domain.Object) (string, error)
	GetObject(ctx context.Context, name string) ([]byte, error)
	DeleteObject(ctx context.Context, name string) error
}

type Server struct {
	storagepb.UnimplementedStorageServer
	addr       string
	grpcServer *grpc.Server
	storage    Storage

	log *slog.Logger
}

func NewServer(addr string, storage Storage, log *slog.Logger) *Server {
	grpcServer := grpc.NewServer()

	server := &Server{
		addr:       addr,
		grpcServer: grpcServer,
		storage:    storage,
		log:        log,
	}

	storagepb.RegisterStorageServer(grpcServer, server)

	return server
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on address %s: %w", s.addr, err)
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

func (s *Server) SaveObject(ctx context.Context, req *storagepb.SaveRequest) (*storagepb.SaveResponse, error) {
	s.log.Debug("save object request")

	obj := domain.Object{
		Name: req.GetObject().GetName(),
		Data: req.GetObject().GetData(),
	}

	etag, err := s.storage.SaveObject(ctx, obj)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error saving object: %v", err)
	}

	return &storagepb.SaveResponse{
		Etag: etag,
	}, nil
}

func (s *Server) GetObject(ctx context.Context, req *storagepb.GetRequest) (*storagepb.GetResponse, error) {
	s.log.Debug("get object request")

	data, err := s.storage.GetObject(ctx, req.GetName())
	if err != nil {
		if errors.Is(err, domain.ErrFileNotFound) {
			return nil, status.Errorf(codes.NotFound, "%v", err)
		}

		return nil, status.Errorf(codes.Internal, "error getting object: %v", err)
	}

	return &storagepb.GetResponse{
		Data: data,
	}, nil
}

func (s *Server) DeleteObject(ctx context.Context, req *storagepb.DeleteRequest) (*emptypb.Empty, error) {
	s.log.Debug("delete object request")

	err := s.storage.DeleteObject(ctx, req.GetName())
	if err != nil {
		if errors.Is(err, domain.ErrFileNotFound) {
			return nil, status.Errorf(codes.NotFound, "%v", err)
		}

		return nil, status.Errorf(codes.Internal, "error deleting object: %v", err)
	}

	return &emptypb.Empty{}, nil
}
