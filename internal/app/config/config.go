package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	A string `env:"SERVER_ADDRESS"`
	B string `env:"BASE_URL"`
	F string `env:"FILE_STORAGE_PATH"`
}

func ConfigInit() *Config {
	flagA := flag.String("a", "localhost:8080", "Адрес сервера")
	flagB := flag.String("b", "http://localhost:8080", "Базовый URL")
	flagF := flag.String("f", "./storage.json", "Путь до файла")
	flag.Parse()

	cfg := Config{
		A: *flagA,
		B: *flagB,
		F: *flagF,
	}

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Ошибка чтения переменных окружения: %v", err)
	}

	return &cfg
}
