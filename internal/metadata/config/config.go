package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type LoggerConfig struct {
	LogLevel string `yaml:"log_level" env:"LOG_LEVEL" env-default:"ERROR"`
}

type DatabaseConfig struct {
	Address       string `yaml:"address" env:"DATABASE_ADDRESS"`
	MigrationsDir string `yaml:"migrations_dir" env:"DATABASE_MIGRATIONS_DIRECTORY"`
}

type GRPCConfig struct {
	Address string `yaml:"address" env:"METADATA_ADDRESS_GRPC" env-default:":50051"`
}

type StorageConfig struct {
	Address string `yaml:"address" env:"STORAGE_ADDRESS"`
}

type Config struct {
	Logger   LoggerConfig   `yaml:"logger"`
	Database DatabaseConfig `yaml:"database"`
	GRPC     GRPCConfig     `yaml:"grpc"`
	Storage  StorageConfig  `yaml:"storage"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
