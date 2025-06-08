package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
)

const (
	DefaultFileStoragePath = "storage.json"
	DefaultServerAddress   = "localhost:8080"
	DefaultBaseURL         = "http://localhost:8080"
	DefaultDatabaseDSN     = ""
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string
	DatabaseDSN     string `env:"DATABASE_DSN" envDefault:"localDB"`
}

func NewConfig() *Config {
	cfg := &Config{}

	// Сначала парсим переменные окружения через caarlos0/env
	if err := env.Parse(cfg); err != nil {
		fmt.Println("Ошибка парсинга переменных окружения:", err)
	}

	// Парсим флаги
	fileStoragePathFlag := flag.String("f", DefaultFileStoragePath, "Путь к файлу хранения данных")
	serverAddressFlag := flag.String("a", DefaultServerAddress, "Адрес сервера (например, localhost:8080)")
	baseURLFlag := flag.String("b", DefaultBaseURL, "Базовый URL для сокращённых ссылок")
	dsnFlag := flag.String("d", DefaultDatabaseDSN, "Строка подключения к базе данных")
	flag.Parse()

	// Определяем итоговые значения по приоритету: env → flags → default
	cfg.FileStoragePath = getConfigValue(os.Getenv("FILE_STORAGE_PATH"), *fileStoragePathFlag, DefaultFileStoragePath)
	cfg.ServerAddress = getConfigValue(os.Getenv("SERVER_ADDRESS"), *serverAddressFlag, DefaultServerAddress)
	cfg.BaseURL = getConfigValue(os.Getenv("BASE_URL"), *baseURLFlag, DefaultBaseURL)
	cfg.DatabaseDSN = getConfigValue(os.Getenv("DATABASE_DSN"), *dsnFlag, DefaultDatabaseDSN)

	return cfg
}

// Хелпер для выбора значения по приоритету
func getConfigValue(envValue, flagValue, defaultValue string) string {
	if envValue != "" {
		return envValue
	}
	if flagValue != "" {
		return flagValue
	}
	return defaultValue
}
