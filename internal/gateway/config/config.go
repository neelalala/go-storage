package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type LoggerConfig struct {
	LogLevel string `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
}

type HTTPConfig struct {
	Address string        `yaml:"address" env:"GATEWAY_ADDRESS_HTTP" env-default:"localhost:80"`
	Timeout time.Duration `yaml:"timeout" env:"GATEWAY_TIMEOUT" env-default:"5s"`
}

type MetadataServiceConfig struct {
	Address string `yaml:"address" env:"METADATA_SERVICE_ADDRESS" env-default:"metadata:50051"`
}

type UsersServiceConfig struct {
	Address string `yaml:"address" env:"USERS_SERVICE_ADDRESS" env-default:"users:50051"`
}

type Config struct {
	Logger          LoggerConfig          `yaml:"logger"`
	HTTP            HTTPConfig            `yaml:"http"`
	MetadataService MetadataServiceConfig `yaml:"metadata"`
	UsersService    UsersServiceConfig    `yaml:"users"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
