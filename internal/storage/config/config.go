package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type LoggerConfig struct {
	LogLevel string `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
}

type GRPCConfig struct {
	Address string `yaml:"address" env:"SERVER_ADDRESS_GRPC" env-default:":50051"`
}

type DiscoveryServiceConfig struct {
	Address           string        `yaml:"address" env:"DISCOVERY_SERVICE_ADDRESS" env-default:"metadata:50051"`
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval" env:"HEARTBEAT_INTERVAL" env-default:"10s"`
}

type NodeConfig struct {
	ID         string `yaml:"id" env:"NODE_ID"`
	UploadRoot string `yaml:"upload_root" env:"STORAGE_UPLOAD_ROOT" env-default:"uploads/"`
}

type Config struct {
	GRPC             GRPCConfig             `yaml:"grpc"`
	Logger           LoggerConfig           `yaml:"logger"`
	DiscoveryService DiscoveryServiceConfig `yaml:"discovery_service"`
	Node             NodeConfig             `yaml:"node"`
}

func Load(configPath string) (Config, error) {
	var cfg Config

	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
	}

	if configPath != "" {
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			return Config{}, fmt.Errorf("read config %q: %w", configPath, err)
		}
		return cfg, nil
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return Config{}, fmt.Errorf("read environment variables: %w", err)
	}

	return cfg, nil
}

func MustLoad(configPath string) Config {
	cfg, err := Load(configPath)
	if err != nil {
		log.Fatalf("failed to load configuration: %s", err)
	}
	return cfg
}
