package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/neelalala/go-storage/internal/gateway/adapter/in/http"
	"github.com/neelalala/go-storage/internal/gateway/adapter/in/http/marshal"
	"github.com/neelalala/go-storage/internal/gateway/adapter/out/grpc/metadata"
	"github.com/neelalala/go-storage/internal/gateway/adapter/out/grpc/storage"
	"github.com/neelalala/go-storage/internal/gateway/application"
	"github.com/neelalala/go-storage/internal/gateway/config"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()

	cfg := config.MustLoad(configPath)

	log := mustMakeLogger(cfg.Logger.LogLevel)

	if err := run(cfg, log); err != nil {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, log *slog.Logger) error {
	log.Info("starting server")
	log.Debug("debug messages are enabled")

	log.Debug("config", fmt.Sprintf("%+v", cfg))

	metadata, err := metadata.New(cfg.MetadataService.Address)
	if err != nil {
		return err
	}

	nodes := storage.NewNodeManager()

	gateway := application.NewGateway(metadata, nodes, log)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	marshaller := marshal.JSONMarshaller{}

	server := http.NewServer(metadata, gateway, marshaller, cfg.HTTP.Address, cfg.HTTP.Timeout, log)

	go func() {
		<-ctx.Done()
		log.Info("shutting down server")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error("error shutting down", "error", err)
		}
	}()

	if err := server.Start(); err != nil {
		return fmt.Errorf("server returned unexpectedly: %w", err)
	}

	return nil
}

func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		panic("unknown log level: " + logLevel)
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level, AddSource: true})
	return slog.New(handler)
}
