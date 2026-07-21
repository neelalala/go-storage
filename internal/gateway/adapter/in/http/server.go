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
	PutObject(ctx context.Context, bucket, key string, data []byte) error
	GetObject(ctx context.Context, bucket, key string) ([]byte, error)
	DeleteObject(ctx context.Context, bucket, key string) error
}

type Marshaller interface {
	ListBuckets(buckets []domain.BucketMetadata) ([]byte, error)
	ListObjectsV2(name, prefix, delimiter string, objects []domain.ObjectMetadata) ([]byte, error)
	Error(err error, resource string, requestID uuid.UUID) ([]byte, int)
}

type Server struct {
	gateway Gateway
	server  *http.Server

	log *slog.Logger
}

func NewServer(
	metadata domain.MetadataService,
	gateway Gateway,
	marshaller Marshaller,
	addr string,
	timeout time.Duration,
	log *slog.Logger,
) *Server {
	handler := NewHandler(metadata, gateway, marshaller, log)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handler.ListBuckets)

	mux.HandleFunc("PUT /{bucket}", handler.CreateBucket)
	mux.HandleFunc("GET /{bucket}", handler.ListObjects)
	mux.HandleFunc("DELETE /{bucket}", handler.DeleteBucket)

	mux.HandleFunc("PUT /{bucket}/{key...}", handler.PutObject)
	mux.HandleFunc("GET /{bucket}/{key...}", handler.GetObject)
	mux.HandleFunc("DELETE /{bucket}/{key...}", handler.DeleteObject)

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
