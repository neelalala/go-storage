package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPConfig struct {
	Address string        `yaml:"address" env:"GATEWAY_ADDRESS_HTTP" env-default:"localhost:80"`
	Timeout time.Duration `yaml:"timeout" env:"GATEWAY_TIMEOUT" env-default:"5s"`
}

type Config struct {
	HTTPConfig     HTTPConfig `yaml:"http"`
	LogLevel       string     `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	StorageAddress string     `yaml:"storage_address" env:"STORAGE_ADDRESS" env-default:"storage:80"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
