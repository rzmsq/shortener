package config

import (
	"log/slog"
	HTTPConfig "shortener/internal/config/http-config"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env                   string        `yaml:"env" required:"true"`
	StoragePath           string        `yaml:"storagePath" required:"true"`
	MaxShutDownTimeServer time.Duration `yaml:"maxShutDownTimeServer" default:"10s"`
	HTTPConfig.HTTPConfig `yaml:"httpConfig" required:"true"`
}

func MustLoadConfig(pathToConfig string) *Config {
	var config Config
	err := cleanenv.ReadConfig(pathToConfig, &config)
	if err != nil {
		slog.Error("Error reading config, use -config [pathToConfig]")
		panic(err)
	}
	return &config
}
