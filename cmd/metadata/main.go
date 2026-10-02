package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neelalala/go-storage/internal/metadata/adapter/in/grpc"
	"github.com/neelalala/go-storage/internal/metadata/adapter/out/grpc/storage"
	"github.com/neelalala/go-storage/internal/metadata/adapter/out/hasher"
	"github.com/neelalala/go-storage/internal/metadata/adapter/out/nodes"
	"github.com/neelalala/go-storage/internal/metadata/adapter/out/repository/sql"
	"github.com/neelalala/go-storage/internal/metadata/application"
	"github.com/neelalala/go-storage/internal/metadata/config"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to configuration file (optional)")
	flag.Parse()

	cfg := config.MustLoad(configPath)

	log := mustMakeLogger(cfg.Logger.LogLevel)

	if err := run(cfg, log); err != nil {
		log.Error("service failed", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, log *slog.Logger) error {
	log.Info("starting service")
	log.Debug("debug messages are enabled")

	log.Debug("config", "value", fmt.Sprintf("%+v", cfg))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.Database.URL)
	if err != nil {
		return err
	}

	transactor := sql.NewTransactor(pool)
	bucketRepo := sql.NewBucketRepository(pool)
	uploadRepo := sql.NewUploadRepository(pool)
	objRepo := sql.NewObjectRepository(pool)

	registry, err := nodes.NewRedisNodeRegistry(ctx, cfg.Redis.URL, cfg.Storage.HeartbeatInterval, cfg.Storage.TTLCountToMarkDead, log)
	if err != nil {
		return fmt.Errorf("failed to initialize redis node registry: %w", err)
	}
	defer func() {
		if err := registry.Close(); err != nil {
			log.Error("error closing redis registry", "error", err)
		}
	}()

	manager := nodes.NewRoundRobinNodeManager(registry)

	hasher := hasher.NewSHA256()

	metadata := application.NewMetadataService(
		transactor,
		bucketRepo,
		uploadRepo,
		objRepo,
		registry,
		manager,
		hasher,
		log,
	)

	gcRepo := sql.NewGCRepository(pool)

	storage := storage.NewNodeManager(registry, log)

	garbageCollector := application.NewGarbageCollector(gcRepo, storage, log)

	server := grpc.NewServer(cfg.Server.Address, metadata, log)

	go func() {
		<-ctx.Done()
		log.Info("shutting down server")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Stop(shutdownCtx); err != nil {
			log.Error("error shutting down", "error", err)
		}
	}()

	go func() {
		log.Info("starting garbage collector")

		garbageCollector.Start(ctx, cfg.GarbageCollector.Interval, cfg.GarbageCollector.TaskLimit, cfg.GarbageCollector.TaskTimeout)
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
