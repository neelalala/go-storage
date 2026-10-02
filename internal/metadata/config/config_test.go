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
	t.Setenv("REDIS_URL", "redis://custom:6379/1")
	t.Setenv("HEARTBEAT_INTERVAL", "15s")
	t.Setenv("TTL_COUNT_TO_MARK_DEAD", "5")
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
	if cfg.Redis.URL != "redis://custom:6379/1" {
		t.Errorf("expected Redis.URL redis://custom:6379/1, got %s", cfg.Redis.URL)
	}
	if cfg.Storage.HeartbeatInterval != 15*time.Second {
		t.Errorf("expected HeartbeatInterval 15s, got %v", cfg.Storage.HeartbeatInterval)
	}
	if cfg.Storage.TTLCountToMarkDead != 5 {
		t.Errorf("expected TTLCountToMarkDead 5, got %d", cfg.Storage.TTLCountToMarkDead)
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
		"REDIS_URL",
		"HEARTBEAT_INTERVAL",
		"TTL_COUNT_TO_MARK_DEAD",
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
	if cfg.Redis.URL != "redis://localhost:6379/0" {
		t.Errorf("expected default Redis.URL redis://localhost:6379/0, got %s", cfg.Redis.URL)
	}
	if cfg.Storage.HeartbeatInterval != 10*time.Second {
		t.Errorf("expected default HeartbeatInterval 10s, got %v", cfg.Storage.HeartbeatInterval)
	}
	if cfg.Storage.TTLCountToMarkDead != 3 {
		t.Errorf("expected default TTLCountToMarkDead 3, got %d", cfg.Storage.TTLCountToMarkDead)
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
redis:
  url: "redis://my-redis:6379/2"
storage:
  heartbeat_interval: 20s
  ttl_count_to_mark_dead: 4
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
	if cfg.Redis.URL != "redis://my-redis:6379/2" {
		t.Errorf("expected Redis.URL redis://my-redis:6379/2, got %s", cfg.Redis.URL)
	}
	if cfg.Storage.HeartbeatInterval != 20*time.Second {
		t.Errorf("expected HeartbeatInterval 20s, got %v", cfg.Storage.HeartbeatInterval)
	}
	if cfg.Storage.TTLCountToMarkDead != 4 {
		t.Errorf("expected TTLCountToMarkDead 4, got %d", cfg.Storage.TTLCountToMarkDead)
	}
}

func TestMetadataConfig_LoadNonExistentFile(t *testing.T) {
	_, err := Load("does_not_exist_config.yaml")
	if err == nil {
		t.Fatalf("expected error when loading non-existent config file, got nil")
	}
}
