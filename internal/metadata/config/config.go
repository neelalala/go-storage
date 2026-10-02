package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type LoggerConfig struct {
	LogLevel string `yaml:"level" env:"LOG_LEVEL" env-default:"DEBUG"`
}

type DatabaseConfig struct {
	URL string `yaml:"url" env:"DATABASE_URL"`
}

type ServerConfig struct {
	Address string `yaml:"address" env:"SERVER_ADDRESS" env-default:":50051"`
}

type RedisConfig struct {
	URL string `yaml:"url" env:"REDIS_URL" env-default:"redis://localhost:6379/0"`
}

type StorageConfig struct {
	HeartbeatInterval  time.Duration `yaml:"heartbeat_interval" env:"HEARTBEAT_INTERVAL" env-default:"10s"`
	TTLCountToMarkDead int           `yaml:"ttl_count_to_mark_dead" env:"TTL_COUNT_TO_MARK_DEAD" env-default:"3"`
}

type GarbageCollectorConfig struct {
	Interval    time.Duration `yaml:"interval" env:"GC_INTERVAL" env-default:"1m"`
	TaskLimit   int           `yaml:"task_limit" env:"GC_TASK_LIMIT" env-default:"100"`
	TaskTimeout time.Duration `yaml:"task_timeout" env:"GC_TASK_TIMEOUT" env-default:"5s"`
}

type Config struct {
	Logger           LoggerConfig           `yaml:"logger"`
	Database         DatabaseConfig         `yaml:"database"`
	Server           ServerConfig           `yaml:"server"`
	Redis            RedisConfig            `yaml:"redis"`
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
