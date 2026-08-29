package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type LoggerConfig struct {
	LogLevel string `yaml:"log_level" env:"LOG_LEVEL" env-default:"ERROR"`
}

type GRPCConfig struct {
	Address string `yaml:"address" env:"SERVER_ADDRESS_GRPC" env-default:":50051"`
}

type DiscoveryServiceConfig struct {
	Address           string        `yaml:"address" env:"DISCOVERY_SERVICE_ADDRESS"`
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval" env:"HEARTBEAT_INTERVAL" env-default:"1m"`
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

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
