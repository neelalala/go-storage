package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUsersConfig_LoadWithoutConfigFile(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	t.Setenv("LOG_LEVEL", "INFO")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/users_db")
	t.Setenv("USERS_ADDRESS_GRPC", ":50070")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.Logger.LogLevel != "INFO" {
		t.Errorf("expected LogLevel INFO, got %s", cfg.Logger.LogLevel)
	}
	if cfg.Database.URL != "postgres://user:pass@localhost:5432/users_db" {
		t.Errorf("expected Database.URL, got %s", cfg.Database.URL)
	}
	if cfg.GRPC.Address != ":50070" {
		t.Errorf("expected Address :50070, got %s", cfg.GRPC.Address)
	}
}

func TestUsersConfig_LoadDefaults(t *testing.T) {
	for _, env := range []string{
		"CONFIG_PATH",
		"LOG_LEVEL",
		"DATABASE_URL",
		"USERS_ADDRESS_GRPC",
		"SERVER_ADDRESS_GRPC",
		"METADATA_ADDRESS_GRPC",
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
	if cfg.GRPC.Address != ":50051" {
		t.Errorf("expected default Address :50051, got %s", cfg.GRPC.Address)
	}
}

func TestUsersConfig_LoadFromFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	yamlContent := `
logger:
  log_level: WARN
grpc:
  address: ":50059"
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
	if cfg.GRPC.Address != ":50059" {
		t.Errorf("expected Address :50059, got %s", cfg.GRPC.Address)
	}
}

func TestUsersConfig_LoadNonExistentFile(t *testing.T) {
	_, err := Load("does_not_exist_config.yaml")
	if err == nil {
		t.Fatalf("expected error when loading non-existent config file, got nil")
	}
}
