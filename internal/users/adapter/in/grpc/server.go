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

	"github.com/neelalala/go-storage/internal/users/application"
	"github.com/neelalala/go-storage/internal/users/domain"
	userspb "github.com/neelalala/go-storage/pkg/proto/users"
)

type Server struct {
	userspb.UnimplementedUsersServer
	addr       string
	grpcServer *grpc.Server
	service    *application.UsersService

	log *slog.Logger
}

func NewServer(addr string, service *application.UsersService, log *slog.Logger) *Server {
	grpcServer := grpc.NewServer()

	server := &Server{
		addr:       addr,
		grpcServer: grpcServer,
		service:    service,
		log:        log,
	}

	userspb.RegisterUsersServer(server.grpcServer, server)

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

func (s *Server) CreateUser(ctx context.Context, req *userspb.CreateUserRequest) (*userspb.User, error) {
	s.log.Debug("create user request")

	name := req.GetDisplayName()

	user, err := s.service.CreateUser(ctx, name)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			s.log.Debug("user already exists",
				"name", name,
			)
			return nil, status.Errorf(codes.AlreadyExists, "failed to create user: %v", err)
		}
		s.log.Error("failed to create user",
			"name", name,
			"error", err,
		)

		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &userspb.User{
		Id:          user.ID.String(),
		DisplayName: user.DisplayName,
	}, nil
}

func (s *Server) GetUserByName(ctx context.Context, req *userspb.GetUserByNameRequest) (*userspb.User, error) {
	s.log.Debug("get user by name request")

	name  := req.GetDisplayName()

	user, err := s.service.GetUserByName(ctx, name)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			s.log.Debug("user does not exist",
				"name", name,
			)
			return nil, status.Errorf(codes.NotFound, "failed to get user by name: %v", err)
		}
		s.log.Error("failed to get user by name",
			"name", name,
			"error", err,
		)
		return nil, status.Errorf(codes.Internal, "failed to get user by name: %v", err)
	}

	return &userspb.User{
		Id: user.ID.String(),
		DisplayName: user.DisplayName,
	}, nil
}