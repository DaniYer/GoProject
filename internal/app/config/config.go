package config

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress   string        `env:"SERVER_ADDRESS"`
	BaseURL         string        `env:"BASE_URL"`
	DatabaseDSN     string        `env:"DATABASE_DSN"`
	FileStoragePath string        `env:"FILE_STORAGE_PATH"`
	SecretKey       string        `env:"SECRET_KEY" envDefault:"supersecretkey"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT" envDefault:"5s"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" envDefault:"5s"`
}

func Load() *Config {
	cfg := &Config{}

	// Сначала читаем флаги:
	flag.StringVar(&cfg.ServerAddress, "a", "", "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", "", "Base URL address")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "File storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database DSN")
	flag.Parse()

	// Затем читаем ENV — перезапишет флаги если они есть
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("Failed to parse env: %v", err)
	}

	// Дефолты:
	if cfg.ServerAddress == "" {
		cfg.ServerAddress = ":8080"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:8080"
	}
	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = "storage.json"
	}

	return cfg
}
