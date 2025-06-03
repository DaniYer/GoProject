package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string
	DatabaseDSN     string `env:"DATABASE_DSN" envDefault:"localDB"`
}

func NewConfig() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		fmt.Println("Ошибка парсинга переменных окружения:", err)
	}
	fileStoragePathFlag := flag.String("f", "storage.json", "Путь к файлу хранения данных")
	serverAddressFlag := flag.String("a", "localhost:8080", "Адрес сервера (например, localhost:8080)")
	baseURLFlag := flag.String("b", "http://localhost:8080", "Базовый URL для сокращённых ссылок")
	dsnFlag := flag.String("d", "localDB", "Строка подключения к базе данных")
	flag.Parse()

	if envPath, exists := os.LookupEnv("FILE_STORAGE_PATH"); exists && envPath != "" {
		cfg.FileStoragePath = envPath
	} else {
		cfg.FileStoragePath = *fileStoragePathFlag
	}
	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = "storage.json"
	}

	if envAddr, exists := os.LookupEnv("SERVER_ADDRESS"); exists && envAddr != "" {
		cfg.ServerAddress = envAddr
	} else {
		cfg.ServerAddress = *serverAddressFlag
	}
	if cfg.ServerAddress == "" {
		cfg.ServerAddress = "localhost:8080"
	}

	if envBase, exists := os.LookupEnv("BASE_URL"); exists && envBase != "" {
		cfg.BaseURL = envBase
	} else {
		cfg.BaseURL = *baseURLFlag
	}

	if envDSN, exists := os.LookupEnv("DATABASE_DSN"); exists && envDSN != "" {
		cfg.DatabaseDSN = envDSN
	} else {
		cfg.DatabaseDSN = *dsnFlag
	}
	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = "localDB"
	}

	return cfg
}

func (c *Config) Print() {
	fmt.Printf("Сервер будет запущен на: %s\n", c.ServerAddress)
	fmt.Printf("Базовый URL для сокращённых ссылок: %s\n", c.BaseURL)
	fmt.Printf("Файл хранения URL: %s\n", c.FileStoragePath)
}
