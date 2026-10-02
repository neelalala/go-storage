package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	var (
		dbURL         string
		migrationsDir string
	)

	flag.StringVar(&dbURL, "db", os.Getenv("DATABASE_URL"), "Database connection URL (defaults to DATABASE_URL)")

	defaultDir := os.Getenv("MIGRATIONS_DIRECTORY")
	if defaultDir == "" {
		defaultDir = "file://migrations/metadata"
	}

	flag.StringVar(&migrationsDir, "dir", defaultDir, "Migrations directory URL or path (defaults to MIGRATIONS_DIRECTORY)")

	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	cmd := "up"
	var cmdArgs []string
	if len(args) > 0 {
		cmd = strings.ToLower(args[0])
		cmdArgs = args[1:]
	}

	if cmd == "help" {
		usage()
		os.Exit(0)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if dbURL == "" {
		logger.Error("database connection URL is required (provide via -db flag or DATABASE_URL environment variable)")
		os.Exit(1)
	}

	migrationsURL := normalizeMigrationsURL(migrationsDir)

	if err := runMigration(dbURL, migrationsURL, cmd, cmdArgs, logger); err != nil {
		logger.Error("migration failed", "command", cmd, "error", err)
		os.Exit(1)
	}
}

func runMigration(dbURL, migrationsURL, cmd string, args []string, log *slog.Logger) error {
	m, err := migrate.New(migrationsURL, dbURL)
	if err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Warn("error closing migration source", "error", srcErr)
		}
		if dbErr != nil {
			log.Warn("error closing migration database", "error", dbErr)
		}
	}()

	switch cmd {
	case "up":
		if len(args) > 0 {
			steps, convErr := strconv.Atoi(args[0])
			if convErr != nil || steps <= 0 {
				return fmt.Errorf("invalid step count for 'up': %q (must be positive integer)", args[0])
			}
			log.Info("applying migrations", "steps", steps)
			err = m.Steps(steps)
		} else {
			log.Info("applying all pending migrations")
			err = m.Up()
		}

		if err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Info("database is up to date, no migrations applied")
				return nil
			}
			return fmt.Errorf("failed to apply migrations: %w", err)
		}
		log.Info("migrations applied successfully")
		return nil

	case "down":
		if len(args) > 0 && args[0] == "all" {
			log.Info("rolling back all migrations")
			err = m.Down()
		} else {
			if len(args) == 0 {
				return errors.New("'down' requires argument: migrate down <steps> or migrate down all")
			}

			var convErr error
			steps, convErr := strconv.Atoi(args[0])
			if convErr != nil || steps <= 0 {
				return fmt.Errorf("invalid step count for 'down': %q (must be positive integer or 'all')", args[0])
			}

			log.Info("rolling back migrations", "steps", steps)
			err = m.Steps(-steps)
		}

		if err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Info("no migrations to roll back")
				return nil
			}
			return fmt.Errorf("failed to roll back migrations: %w", err)
		}
		log.Info("rollback completed successfully")
		return nil

	case "version":
		ver, dirty, err := m.Version()
		if err != nil {
			if errors.Is(err, migrate.ErrNilVersion) {
				log.Info("no migrations applied yet", "version", "nil")
				return nil
			}
			return fmt.Errorf("failed to get migration version: %w", err)
		}
		log.Info("current migration status", "version", ver, "dirty", dirty)
		if dirty {
			log.Warn("database schema is in dirty state! Use 'force <version>' after resolving issues")
		}
		return nil

	case "force":
		if len(args) == 0 {
			return errors.New("'force' requires version argument: migrate force <version>")
		}
		ver, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid version for 'force': %q (must be integer)", args[0])
		}
		log.Info("forcing migration version", "version", ver)
		if err := m.Force(ver); err != nil {
			return fmt.Errorf("failed to force version %d: %w", ver, err)
		}
		log.Info("migration version forced successfully", "version", ver)
		return nil

	default:
		return fmt.Errorf("unknown command %q (supported: up, down, version, force)", cmd)
	}
}

func normalizeMigrationsURL(dir string) string {
	if strings.Contains(dir, "://") {
		return dir
	}
	return "file://" + dir
}

func usage() {
	fmt.Fprintf(os.Stderr, `go-storage database migration tool

Usage:
  migrate [flags] <command> [arguments]

Commands:
  up [N]         Apply all pending migrations (or N steps if specified)
  down [N|all]   Roll back 1 migration (or N steps, or 'all')
  version        Print current migration version and dirty state
  force <ver>    Set migration version without running migration (clears dirty state)
  help           Show this help message

Flags:
  -db string     Database URL (defaults to DATABASE_URL environment variable)
  -dir string    Migrations directory URL (defaults to MIGRATIONS_DIRECTORY, default: file://migrations/metadata)
`)
}
