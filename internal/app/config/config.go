package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	A string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	B string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	F string `env:"FILE_STORAGE_PATH" envDefault:"./storage.json"`
	D string `env:"DATABASE_DSN" envDefault:"host=localhost user=postgres password=sqwaed21 dbname=urls sslmode=disable"`
}

func ConfigInit() *Config {
	cfg := Config{}

	// Сначала читаем переменные окружения
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Ошибка чтения переменных окружения: %v", err)
	}

	// Потом уже парсим флаги
	flag.StringVar(&cfg.A, "a", cfg.A, "Адрес сервера")
	flag.StringVar(&cfg.B, "b", cfg.B, "Базовый URL")
	flag.StringVar(&cfg.F, "f", cfg.F, "Путь до файла")
	flag.StringVar(&cfg.D, "d", cfg.D, "DSN для базы данных")
	flag.Parse()

	return &cfg
}
