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

type DatabaseConfig struct {
	URL           string `yaml:"url" env:"DATABASE_URL"`
	MigrationsDir string `yaml:"migrations_dir" env:"DATABASE_MIGRATIONS_DIRECTORY" env-default:"file://migrations/metadata"`
}

type GRPCConfig struct {
	Address string `yaml:"address" env:"METADATA_ADDRESS_GRPC" env-default:":50051"`
}

type StorageConfig struct {
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval" env:"HEARTBEAT_INTERVAL" env-default:"10s"`
}

type GarbageCollectorConfig struct {
	Interval    time.Duration `yaml:"interval" env:"GC_INTERVAL" env-default:"1m"`
	TaskLimit   int           `yaml:"task_limit" env:"GC_TASK_LIMIT" env-default:"100"`
	TaskTimeout time.Duration `yaml:"task_timeout" env:"GC_TASK_TIMEOUT" env-default:"5s"`
}

type Config struct {
	Logger           LoggerConfig           `yaml:"logger"`
	Database         DatabaseConfig         `yaml:"database"`
	GRPC             GRPCConfig             `yaml:"grpc"`
	Storage          StorageConfig          `yaml:"storage"`
	GarbageCollector GarbageCollectorConfig `yaml:"garbage_collector"`
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
