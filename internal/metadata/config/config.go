package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type LoggerConfig struct {
	LogLevel string `yaml:"log_level" env:"LOG_LEVEL" env-default:"ERROR"`
}

type DatabaseConfig struct {
	URL           string `yaml:"url" env:"DATABASE_URL"`
	MigrationsDir string `yaml:"migrations_dir" env:"DATABASE_MIGRATIONS_DIRECTORY"`
}

type GRPCConfig struct {
	Address string `yaml:"address" env:"METADATA_ADDRESS_GRPC" env-default:":50051"`
}

type StorageConfig struct {
	ID      string `yaml:"id"`
	Address string `yaml:"address"`
}

type GarbageCollectorConfig struct {
	Interval    time.Duration `yaml:"interval" env:"GC_INTERVAL" env-default:"1m"`
	TaskLimit   int           `yaml:"task_limit" env:"GC_TASK_LIMIT" env-default:"50"`
	TaskTimeout time.Duration `yaml:"task_timeout" env:"GC_TASK_TIMEOUT" env-default:"10s"`
}

type Config struct {
	Logger           LoggerConfig           `yaml:"logger"`
	Database         DatabaseConfig         `yaml:"database"`
	GRPC             GRPCConfig             `yaml:"grpc"`
	Storage          StorageConfig          `yaml:"storage"`
	GarbageCollector GarbageCollectorConfig `yaml:"garbage_collector"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
