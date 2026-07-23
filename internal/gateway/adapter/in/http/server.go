package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/neelalala/go-storage/internal/gateway/adapter/in/http/middleware"
	"github.com/neelalala/go-storage/internal/gateway/application"
	"github.com/neelalala/go-storage/internal/gateway/domain"
)

type Marshaller interface {
	ListBuckets(owner domain.User, buckets []domain.BucketMetadata) ([]byte, error)
	ListObjectsV2(name, prefix, delimiter string, limit int, objects []domain.ObjectMetadata, prefixes []string, isTruncated bool) ([]byte, error)
	Error(err error, resource string, requestID uuid.UUID) ([]byte, int)
}

type Server struct {
	gateway *application.Gateway
	server  *http.Server

	log *slog.Logger
}

func NewServer(
	gateway *application.Gateway,
	marshaller Marshaller,
	addr string,
	timeout time.Duration,
	verifier middleware.Verifier,
	log *slog.Logger,
) *Server {
	handler := NewHandler(gateway, marshaller, log)

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /users/{username}", middleware.RequestID(handler.CreateUser))

	mux.HandleFunc("GET /storage/", middleware.RequestID(middleware.Auth(handler.ListBuckets, verifier)))
	mux.HandleFunc("PUT /storage/{bucket}", middleware.RequestID(middleware.Auth(handler.CreateBucket, verifier)))
	mux.HandleFunc("HEAD /storage/{bucket}", middleware.RequestID(middleware.Auth(handler.HeadBucket, verifier)))
	mux.HandleFunc("DELETE /storage/{bucket}", middleware.RequestID(middleware.Auth(handler.DeleteBucket, verifier)))

	mux.HandleFunc("GET /storage/{bucket}", middleware.RequestID(middleware.Auth(handler.ListObjects, verifier)))
	mux.HandleFunc("PUT /storage/{bucket}/{key...}", middleware.RequestID(middleware.Auth(handler.PutObject, verifier)))
	mux.HandleFunc("HEAD /storage/{bucket}/{key...}", middleware.RequestID(middleware.Auth(handler.HeadObject, verifier)))
	mux.HandleFunc("GET /storage/{bucket}/{key...}", middleware.RequestID(middleware.Auth(handler.GetObject, verifier)))
	mux.HandleFunc("DELETE /storage/{bucket}/{key...}", middleware.RequestID(middleware.Auth(handler.DeleteObject, verifier)))

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
