package config

import (
	"flag"
	"fmt"
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	A string `env:"SERVER_ADDRESS"`
	B string `env:"BASE_URL"`
	F string `env:"FILE_STORAGE_PATH"`
	D string `env:"DATABASE_DSN"`
}

func ConfigInit() *Config {

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		"localhost", "videos", "userpassword", "urls")
	flagA := flag.String("a", "localhost:8080", "Адрес сервера")
	flagB := flag.String("b", "http://localhost:8080", "Базовый URL")
	flagF := flag.String("f", "./storage.json", "Путь до файла")
	flagD := flag.String("d", dsn, "DSN для базы данных")

	flag.Parse()
	//add
	cfg := Config{
		A: *flagA,
		B: *flagB,
		F: *flagF,
		D: *flagD,
	}

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Ошибка чтения переменных окружения: %v", err)
	}

	return &cfg
}
