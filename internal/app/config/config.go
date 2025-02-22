package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string
}

func NewConfig() *Config {
	cfg := &Config{}

	// Сначала пытаемся прочитать переменные окружения
	if err := env.Parse(cfg); err != nil {
		fmt.Println("Ошибка парсинга переменных окружения:", err)
	}

	// Определяем флаги для всех параметров
	// Значения по умолчанию берутся из структуры (с envDefault), но если флаг передан – он будет использован
	fileStoragePathFlag := flag.String("f", "storage.json", "Путь к файлу хранения данных")
	serverAddressFlag := flag.String("a", "localhost:8080", "Адрес сервера (например, localhost:8080)")
	baseURLFlag := flag.String("b", "http://localhost:8080", "Базовый URL для сокращённых ссылок")

	flag.Parse()

	// Для каждого параметра, если переменная окружения задана, она имеет приоритет
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
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:8080"
	}

	return cfg
}

func (c *Config) Print() {
	fmt.Printf("Сервер будет запущен на: %s\n", c.ServerAddress)
	fmt.Printf("Базовый URL для сокращённых ссылок: %s\n", c.BaseURL)
	fmt.Printf("Файл хранения URL: %s\n", c.FileStoragePath)
}
