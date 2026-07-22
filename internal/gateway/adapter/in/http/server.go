package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/neelalala/go-storage/internal/gateway/domain"
)

type Gateway interface {
	CreateUser(ctx context.Context, name string) (domain.User, error)
	ListBuckets(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.BucketMetadata, error)
	CreateBucket(ctx context.Context, userID uuid.UUID, name string) (domain.BucketMetadata, error)
	ListObjects(ctx context.Context, userID uuid.UUID, bucket, prefix, delimiter string, limit, offset int) ([]domain.ObjectMetadata, error)
	DeleteBucket(ctx context.Context, userID uuid.UUID, name string) error
	PutObject(ctx context.Context, userID uuid.UUID, bucket string, key string, data []byte, contentType string, systemMetadata, userMetadata map[string]string) error
	GetObject(ctx context.Context, userID uuid.UUID, bucket, key string) (domain.ObjectMetadata, []byte, error)
	DeleteObject(ctx context.Context, userID uuid.UUID, bucket, key string) error
}

type Marshaller interface {
	ListBuckets(owner domain.User, buckets []domain.BucketMetadata) ([]byte, error)
	ListObjectsV2(name, prefix, delimiter string, objects []domain.ObjectMetadata) ([]byte, error)
	Error(err error, resource string, requestID uuid.UUID) ([]byte, int)
}

type Server struct {
	gateway Gateway
	server  *http.Server

	log *slog.Logger
}

func NewServer(
	gateway Gateway,
	marshaller Marshaller,
	addr string,
	timeout time.Duration,
	log *slog.Logger,
) *Server {
	handler := NewHandler(gateway, marshaller, log)

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /users", handler.CreateUser)

	mux.HandleFunc("GET /storage/", handler.ListBuckets)

	mux.HandleFunc("PUT /storage/{bucket}", handler.CreateBucket)
	mux.HandleFunc("GET /storage/{bucket}", handler.ListObjects)
	mux.HandleFunc("DELETE /storage/{bucket}", handler.DeleteBucket)

	mux.HandleFunc("PUT /storage/{bucket}/{key...}", handler.PutObject)
	mux.HandleFunc("GET /storage/{bucket}/{key...}", handler.GetObject)
	mux.HandleFunc("DELETE /storage/{bucket}/{key...}", handler.DeleteObject)

	server := &http.Server{
		Addr:        addr,
		ReadTimeout: timeout,
		Handler:     mux,
	}

	return &Server{
		gateway: gateway,
		server:  server,
		log:     log,
	}
}

func (s *Server) Start() error {
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
