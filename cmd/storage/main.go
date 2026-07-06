package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/neelalala/go-storage/internal/storage/adapter/in/grpc"
	"github.com/neelalala/go-storage/internal/storage/adapter/out/store"
	"github.com/neelalala/go-storage/internal/storage/application"
	"github.com/neelalala/go-storage/internal/storage/config"
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

	store := store.New(cfg.UploadRoot)

	storage := application.NewStorage(store, cfg.NodeName, log)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	server := grpc.NewServer(cfg.GRPC.Address, storage, log)

	go func() {
		<-ctx.Done()
		log.Info("shutting down server")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Stop(shutdownCtx); err != nil {
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
