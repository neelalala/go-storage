package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGatewayConfig_LoadWithoutConfigFile(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	t.Setenv("LOG_LEVEL", "INFO")
	t.Setenv("GATEWAY_ADDRESS_HTTP", ":9090")
	t.Setenv("METADATA_SERVICE_ADDRESS", "custom-meta:50051")
	t.Setenv("USERS_SERVICE_ADDRESS", "custom-users:50051")
	t.Setenv("GATEWAY_TIMEOUT", "10s")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.Logger.LogLevel != "INFO" {
		t.Errorf("expected LogLevel INFO, got %s", cfg.Logger.LogLevel)
	}
	if cfg.HTTP.Address != ":9090" {
		t.Errorf("expected Address :9090, got %s", cfg.HTTP.Address)
	}
	if cfg.HTTP.Timeout != 10*time.Second {
		t.Errorf("expected Timeout 10s, got %v", cfg.HTTP.Timeout)
	}
	if cfg.MetadataService.Address != "custom-meta:50051" {
		t.Errorf("expected MetadataService.Address custom-meta:50051, got %s", cfg.MetadataService.Address)
	}
	if cfg.UsersService.Address != "custom-users:50051" {
		t.Errorf("expected UsersService.Address custom-users:50051, got %s", cfg.UsersService.Address)
	}
}

func TestGatewayConfig_LoadDefaults(t *testing.T) {
	for _, env := range []string{
		"CONFIG_PATH",
		"LOG_LEVEL",
		"GATEWAY_ADDRESS_HTTP",
		"HTTP_ADDRESS",
		"GATEWAY_TIMEOUT",
		"HTTP_TIMEOUT",
		"METADATA_SERVICE_ADDRESS",
		"USERS_SERVICE_ADDRESS",
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
	if cfg.HTTP.Address != ":80" {
		t.Errorf("expected default Address :80, got %s", cfg.HTTP.Address)
	}
	if cfg.HTTP.Timeout != 5*time.Second {
		t.Errorf("expected default Timeout 5s, got %v", cfg.HTTP.Timeout)
	}
	if cfg.MetadataService.Address != "metadata:50051" {
		t.Errorf("expected default MetadataService.Address metadata:50051, got %s", cfg.MetadataService.Address)
	}
	if cfg.UsersService.Address != "users:50051" {
		t.Errorf("expected default UsersService.Address users:50051, got %s", cfg.UsersService.Address)
	}
}

func TestGatewayConfig_LoadFromFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	yamlContent := `
logger:
  log_level: WARN
http:
  address: ":9999"
  timeout: 3s
metadata:
  address: "m:50051"
users:
  address: "u:50051"
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
	if cfg.HTTP.Address != ":9999" {
		t.Errorf("expected Address :9999, got %s", cfg.HTTP.Address)
	}

	t.Setenv("GATEWAY_ADDRESS_HTTP", ":7777")
	cfgOverridden, err := Load(configPath)
	if err != nil {
		t.Fatalf("unexpected error loading config with env override: %v", err)
	}
	if cfgOverridden.HTTP.Address != ":7777" {
		t.Errorf("expected env override :7777, got %s", cfgOverridden.HTTP.Address)
	}
}

func TestGatewayConfig_LoadNonExistentFile(t *testing.T) {
	_, err := Load("does_not_exist_config.yaml")
	if err == nil {
		t.Fatalf("expected error when loading non-existent config file, got nil")
	}
}
