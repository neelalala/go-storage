package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neelalala/go-storage/internal/users/adapter/in/grpc"
	"github.com/neelalala/go-storage/internal/users/adapter/out/repository/sql"
	"github.com/neelalala/go-storage/internal/users/application"
	"github.com/neelalala/go-storage/internal/users/config"
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

	userRepo := sql.NewUserRepository(pool)

	users := application.NewUsersService(userRepo)
	server := grpc.NewServer(cfg.GRPC.Address, users, log)

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
