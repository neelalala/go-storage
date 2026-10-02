package config

import (
	"fmt"
	"log"
	"os"

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

type Config struct {
	Logger   LoggerConfig   `yaml:"logger"`
	Database DatabaseConfig `yaml:"database"`
	Server   ServerConfig   `yaml:"server"`
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
