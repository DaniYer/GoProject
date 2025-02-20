package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env"
)

// Config структура для хранения конфигурации
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"` // ✅ Без http://
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string
}

// NewConfig функция для создания и парсинга конфигурации
func NewConfig() *Config {
	cfg := &Config{}

	// Парсим переменные окружения
	err := env.Parse(cfg)
	if err != nil {
		fmt.Println("Ошибка парсинга переменных окружения:", err)
	}

	// Определяем путь к файлу хранения
	defaultPath := "storage.json"

	// Если установлена переменная окружения, используем её
	if envPath, exists := os.LookupEnv("FILE_STORAGE_PATH"); exists {
		cfg.FileStoragePath = envPath
	} else {
		// Иначе проверяем переданный флаг
		flag.StringVar(&cfg.FileStoragePath, "f", defaultPath, "Путь к файлу хранения данных")
		flag.Parse()
	}

	return cfg
}

// Вывод конфигурации (для удобства)
func (c *Config) Print() {
	fmt.Printf("Сервер будет запущен на: %s\n", c.ServerAddress)
	fmt.Printf("Базовый URL для сокращённых ссылок: %s\n", c.BaseURL)
	fmt.Printf("Файл хранения URL: %s\n", c.FileStoragePath)
}
