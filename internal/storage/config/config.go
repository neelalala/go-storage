package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type GRPCConfig struct {
	Address string `yaml:"address" env:"STORAGE_ADDRESS_GRPC" env-default:":50051"`
}

type Config struct {
	GRPCConfig GRPCConfig `yaml:"grpc"`
	LogLevel   string     `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	UploadRoot string     `yaml:"upload_root" env:"STORAGE_UPLOAD_ROOT" env-default:"uploads/"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
