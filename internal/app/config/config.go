package config

import (
	"flag"
	"fmt"
	"log"

	"github.com/caarlos0/env"
)

// Config структура для хранения конфигурации
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"storage.json"`
}

// NewConfig функция для создания и парсинга конфигурации
func NewConfig() *Config {
	cfg := &Config{}

	// Парсим флаги командной строки
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Адрес для запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Базовый адрес для сокращённых URL")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "Путь до файла хранения URL")
	flag.Parse()

	// Парсим переменные окружения
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("Ошибка при парсинге переменных окружения: %v", err)
	}

	return cfg
}

// Вывод конфигурации (для удобства)
func (c *Config) Print() {
	fmt.Printf("Сервер будет запущен на: %s\n", c.ServerAddress)
	fmt.Printf("Базовый URL для сокращённых ссылок: %s\n", c.BaseURL)
	fmt.Printf("Путь до файла хранения URL: %s\n", c.FileStoragePath)
}
