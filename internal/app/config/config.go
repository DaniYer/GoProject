package config

import (
	"flag"
	"fmt"
)

// Config структура для хранения конфигурации
type Config struct {
	ServerAddress string
	BaseURL       string
}

// NewConfig функция для создания и парсинга конфигурации
func NewConfig() *Config {
	cfg := &Config{}
	// Парсинг флагов
	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "Адрес для запуска HTTP-сервера (например, localhost:8888)")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Базовый адрес для сокращённых URL (например, http://localhost:8000)")
	flag.Parse()
	return cfg
}

// Вывод конфигурации (для удобства)
func (c *Config) Print() {
	fmt.Printf("Сервер будет запущен на: %s\n", c.ServerAddress)
	fmt.Printf("Базовый URL для сокращённых ссылок: %s\n", c.BaseURL)
}
