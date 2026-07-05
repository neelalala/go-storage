package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
)

type Gateway interface {
	PutObject(ctx context.Context, bucket, key string, data []byte) error
	GetObject(ctx context.Context, bucket, key string) ([]byte, error)
	DeleteObject(ctx context.Context, bucket, key string) error
}

type Server struct {
	gateway Gateway
	server  *http.Server

	log *slog.Logger
}

func NewServer(gateway Gateway, addr string, log *slog.Logger) *Server {
	handler := NewHandler(gateway, log)

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /{bucket}/{key...}", handler.PutObject)
	mux.HandleFunc("GET /{bucket}/{key...}", handler.GetObject)
	mux.HandleFunc("DELETE /{bucket}/{key...}", handler.DeleteObject)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
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
