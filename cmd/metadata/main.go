package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neelalala/go-storage/internal/metadata/adapter/in/grpc"
	"github.com/neelalala/go-storage/internal/metadata/adapter/out/grpc/storage"
	"github.com/neelalala/go-storage/internal/metadata/adapter/out/hasher"
	"github.com/neelalala/go-storage/internal/metadata/adapter/out/migrations"
	"github.com/neelalala/go-storage/internal/metadata/adapter/out/repository/sql"
	"github.com/neelalala/go-storage/internal/metadata/application"
	"github.com/neelalala/go-storage/internal/metadata/config"
	"github.com/neelalala/go-storage/internal/metadata/domain"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "service configuration file")
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

	log.Debug("config", fmt.Sprintf("%+v", cfg))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.Database.URL)
	if err != nil {
		return err
	}

	err = migrations.RunMigrationsFromFile(cfg.Database.URL, cfg.Database.MigrationsDir, log)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	bucketRepo := sql.NewBucketRepository(pool)
	uploadRepo := sql.NewUploadRepository(pool)
	objRepo := sql.NewObjectRepository(pool)

	storageUUID, err := uuid.Parse(cfg.Storage.ID)
	if err != nil {
		return err
	}

	hasher := hasher.NewSHA256()

	metadata := application.NewMetadataService(
		bucketRepo,
		uploadRepo,
		objRepo,
		domain.Storage{
			ID:      storageUUID,
			Address: cfg.Storage.Address,
		},
		hasher,
		log,
	)

	storage, err := storage.New(cfg.Storage.Address)
	if err != nil {
		return err
	}

	gcRepo := sql.NewGCRepository(pool)

	garbageCollector := application.NewGarbageCollector(gcRepo, storage, log)

	server := grpc.NewServer(cfg.GRPC.Address, metadata, log)

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
