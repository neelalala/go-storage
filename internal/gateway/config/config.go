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

type ServerConfig struct {
	Address string        `yaml:"address" env:"SERVER_ADDRESS" env-default:":80"`
	Timeout time.Duration `yaml:"timeout" env:"SERVER_TIMEOUT" env-default:"5s"`
}

type MetadataServiceConfig struct {
	Address string `yaml:"address" env:"METADATA_SERVICE_ADDRESS" env-default:"metadata:50051"`
}

type UsersServiceConfig struct {
	Address string `yaml:"address" env:"USERS_SERVICE_ADDRESS" env-default:"users:50051"`
}

type Config struct {
	Logger          LoggerConfig          `yaml:"logger"`
	Server          ServerConfig          `yaml:"server"`
	MetadataService MetadataServiceConfig `yaml:"metadata"`
	UsersService    UsersServiceConfig    `yaml:"users"`
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
