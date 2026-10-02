package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMetadataConfig_LoadWithoutConfigFile(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	t.Setenv("LOG_LEVEL", "INFO")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	t.Setenv("SERVER_ADDRESS", ":50060")
	t.Setenv("HEARTBEAT_INTERVAL", "15s")
	t.Setenv("GC_INTERVAL", "2m")
	t.Setenv("GC_TASK_LIMIT", "200")
	t.Setenv("GC_TASK_TIMEOUT", "8s")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.Logger.LogLevel != "INFO" {
		t.Errorf("expected LogLevel INFO, got %s", cfg.Logger.LogLevel)
	}
	if cfg.Database.URL != "postgres://user:pass@localhost:5432/db" {
		t.Errorf("expected Database.URL, got %s", cfg.Database.URL)
	}
	if cfg.Database.MigrationsDir != "file://migrations/metadata" {
		t.Errorf("expected default MigrationsDir file://migrations/metadata, got %s", cfg.Database.MigrationsDir)
	}
	if cfg.Server.Address != ":50060" {
		t.Errorf("expected Address :50060, got %s", cfg.Server.Address)
	}
	if cfg.Storage.HeartbeatInterval != 15*time.Second {
		t.Errorf("expected HeartbeatInterval 15s, got %v", cfg.Storage.HeartbeatInterval)
	}
	if cfg.GarbageCollector.Interval != 2*time.Minute {
		t.Errorf("expected GC Interval 2m, got %v", cfg.GarbageCollector.Interval)
	}
	if cfg.GarbageCollector.TaskLimit != 200 {
		t.Errorf("expected GC TaskLimit 200, got %d", cfg.GarbageCollector.TaskLimit)
	}
	if cfg.GarbageCollector.TaskTimeout != 8*time.Second {
		t.Errorf("expected GC TaskTimeout 8s, got %v", cfg.GarbageCollector.TaskTimeout)
	}
}

func TestMetadataConfig_LoadDefaults(t *testing.T) {
	for _, env := range []string{
		"CONFIG_PATH",
		"LOG_LEVEL",
		"DATABASE_URL",
		"MIGRATIONS_DIRECTORY",
		"SERVER_ADDRESS",
		"HEARTBEAT_INTERVAL",
		"GC_INTERVAL",
		"GC_TASK_LIMIT",
		"GC_TASK_TIMEOUT",
	} {
		os.Unsetenv(env)
	}

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.Logger.LogLevel != "DEBUG" {
		t.Errorf("expected default LogLevel DEBUG, got %s", cfg.Logger.LogLevel)
	}
	if cfg.Database.MigrationsDir != "file://migrations/metadata" {
		t.Errorf("expected default MigrationsDir file://migrations/metadata, got %s", cfg.Database.MigrationsDir)
	}
	if cfg.Server.Address != ":50051" {
		t.Errorf("expected default Address :50051, got %s", cfg.Server.Address)
	}
	if cfg.Storage.HeartbeatInterval != 10*time.Second {
		t.Errorf("expected default HeartbeatInterval 10s, got %v", cfg.Storage.HeartbeatInterval)
	}
	if cfg.GarbageCollector.Interval != 1*time.Minute {
		t.Errorf("expected default Interval 1m, got %v", cfg.GarbageCollector.Interval)
	}
	if cfg.GarbageCollector.TaskLimit != 100 {
		t.Errorf("expected default TaskLimit 100, got %d", cfg.GarbageCollector.TaskLimit)
	}
	if cfg.GarbageCollector.TaskTimeout != 5*time.Second {
		t.Errorf("expected default TaskTimeout 5s, got %v", cfg.GarbageCollector.TaskTimeout)
	}
}

func TestMetadataConfig_LoadFromFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	yamlContent := `
logger:
  level: WARN
server:
  address: ":50055"
storage:
  heartbeat_interval: 20s
garbage_collector:
  interval: 5m
  task_limit: 50
  task_timeout: 10s
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("unexpected error loading config from file: %v", err)
	}

	if cfg.Logger.LogLevel != "WARN" {
		t.Errorf("expected LogLevel WARN, got %s", cfg.Logger.LogLevel)
	}
	if cfg.Server.Address != ":50055" {
		t.Errorf("expected Address :50055, got %s", cfg.Server.Address)
	}
	if cfg.Storage.HeartbeatInterval != 20*time.Second {
		t.Errorf("expected HeartbeatInterval 20s, got %v", cfg.Storage.HeartbeatInterval)
	}
}

func TestMetadataConfig_LoadNonExistentFile(t *testing.T) {
	_, err := Load("does_not_exist_config.yaml")
	if err == nil {
		t.Fatalf("expected error when loading non-existent config file, got nil")
	}
}
