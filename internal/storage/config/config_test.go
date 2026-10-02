package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStorageConfig_LoadWithoutConfigFile(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	t.Setenv("LOG_LEVEL", "INFO")
	t.Setenv("SERVER_ADDRESS_GRPC", "node1:50051")
	t.Setenv("DISCOVERY_SERVICE_ADDRESS", "custom-meta:50051")
	t.Setenv("HEARTBEAT_INTERVAL", "15s")
	t.Setenv("NODE_ID", "test-node-uuid")
	t.Setenv("STORAGE_UPLOAD_ROOT", "/data/uploads")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.Logger.LogLevel != "INFO" {
		t.Errorf("expected LogLevel INFO, got %s", cfg.Logger.LogLevel)
	}
	if cfg.GRPC.Address != "node1:50051" {
		t.Errorf("expected Address node1:50051, got %s", cfg.GRPC.Address)
	}
	if cfg.DiscoveryService.Address != "custom-meta:50051" {
		t.Errorf("expected DiscoveryService Address custom-meta:50051, got %s", cfg.DiscoveryService.Address)
	}
	if cfg.DiscoveryService.HeartbeatInterval != 15*time.Second {
		t.Errorf("expected HeartbeatInterval 15s, got %v", cfg.DiscoveryService.HeartbeatInterval)
	}
	if cfg.Node.ID != "test-node-uuid" {
		t.Errorf("expected Node.ID test-node-uuid, got %s", cfg.Node.ID)
	}
	if cfg.Node.UploadRoot != "/data/uploads" {
		t.Errorf("expected Node.UploadRoot /data/uploads, got %s", cfg.Node.UploadRoot)
	}
}

func TestStorageConfig_LoadDefaults(t *testing.T) {
	for _, env := range []string{
		"CONFIG_PATH",
		"LOG_LEVEL",
		"SERVER_ADDRESS_GRPC",
		"GRPC_ADDRESS",
		"DISCOVERY_SERVICE_ADDRESS",
		"HEARTBEAT_INTERVAL",
		"NODE_ID",
		"STORAGE_UPLOAD_ROOT",
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
	if cfg.DiscoveryService.Address != "metadata:50051" {
		t.Errorf("expected default DiscoveryService Address metadata:50051, got %s", cfg.DiscoveryService.Address)
	}
	if cfg.DiscoveryService.HeartbeatInterval != 10*time.Second {
		t.Errorf("expected default HeartbeatInterval 10s, got %v", cfg.DiscoveryService.HeartbeatInterval)
	}
	if cfg.Node.UploadRoot != "uploads/" {
		t.Errorf("expected default UploadRoot uploads/, got %s", cfg.Node.UploadRoot)
	}
}

func TestStorageConfig_LoadFromFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	yamlContent := `
logger:
  log_level: WARN
grpc:
  address: ":50052"
node:
  upload_root: /custom/uploads
discovery_service:
  address: "meta:50051"
  heartbeat_interval: 5s
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
	if cfg.GRPC.Address != ":50052" {
		t.Errorf("expected Address :50052, got %s", cfg.GRPC.Address)
	}
	if cfg.Node.UploadRoot != "/custom/uploads" {
		t.Errorf("expected UploadRoot /custom/uploads, got %s", cfg.Node.UploadRoot)
	}
}

func TestStorageConfig_LoadNonExistentFile(t *testing.T) {
	_, err := Load("does_not_exist_config.yaml")
	if err == nil {
		t.Fatalf("expected error when loading non-existent config file, got nil")
	}
}
