// Package initapp отвечает за инициализацию и запуск всего приложения.
// Здесь собираются конфигурация, логирование, хранилища, сервисы, маршруты и middlewares.
package initapp

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	_ "github.com/DaniYer/GoProject.git/api/docs" // импортируем для генерации Swagger документации
	"github.com/DaniYer/GoProject.git/internal/app/config"
	"github.com/DaniYer/GoProject.git/internal/app/handlers"
	"github.com/DaniYer/GoProject.git/internal/app/middlewares"
	"github.com/DaniYer/GoProject.git/internal/app/service"
	"github.com/DaniYer/GoProject.git/internal/app/storage/database"
	"github.com/DaniYer/GoProject.git/internal/app/storage/file"
	"github.com/DaniYer/GoProject.git/internal/app/storage/memory"
	"github.com/DaniYer/GoProject.git/internal/app/worker"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL драйвер
	"github.com/pressly/goose"         // Миграции
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

// InitializeApp инициализирует конфигурацию, логирование, подключение к базе данных,
// хранилища, сервисы, роуты, middlewares и запускает HTTP-сервер.
// Возвращает ошибку, если запуск не удался.
func InitializeApp() error {
	// Загружаем конфигурацию приложения (параметры сервера, DSN, пути к файлам и т.д.)
	cfg := config.NewConfig()

	// Инициализация логгера
	logger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("init logger error: %w", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()
	middlewares.InitLogger(sugar)

	var (
		db    *sql.DB
		store service.URLStore
	)

	// Если указан DSN базы данных — подключаем PostgreSQL и запускаем миграции
	if cfg.DatabaseDSN != "" && cfg.DatabaseDSN != config.DefaultDatabaseDSN {
		db, err = database.InitDB("pgx", cfg.DatabaseDSN)
		if err != nil {
			sugar.Errorf("DB connect error: %v", err)
			return err
		}
		if err := goose.Up(db, "internal/app/storage/database/migrations"); err != nil {
			sugar.Errorf("Migration error: %v", err)
			return err
		}
		store = database.NewDBStore(db)
	}

	// Если БД нет, пробуем файловое хранилище
	if store == nil && cfg.FileStoragePath != "" {
		fs, err := file.NewFileStore(cfg.FileStoragePath)
		if err != nil {
			sugar.Errorf("FileStore init error: %v", err)
		} else {
			store = fs
		}
	}

	// Если и файлового нет — используем in-memory
	if store == nil {
		sugar.Infof("Using in-memory storage")
		store = memory.NewMemoryStore()
	}

	// Сервис работы с короткими ссылками
	urlService := service.NewURLService(store, cfg.BaseURL)

	// Запускаем пул воркеров для асинхронного удаления
	workerPool := worker.NewDeleteWorkerPool(urlService, 1024, 5*time.Second)
	workerPool.Start()

	// Создаём роутер
	router := chi.NewRouter()

	sugar.Infow("Start server", "addr", cfg.ServerAddress)

	// Подключаем middlewares
	router.Use(middlewares.WithLogging)    // Логирование запросов
	router.Use(middlewares.GzipHandle)     // Сжатие gzip
	router.Use(middlewares.AuthMiddleware) // Авторизация через cookie

	// Регистрация маршрутов
	router.Post("/", handlers.NewGenerateShortURLHandler(urlService))
	router.Post("/api/shorten", handlers.NewHandleShortenURLv13(urlService))
	router.Post("/api/shorten/batch", handlers.NewBatchShortenURLHandler(urlService))
	router.Get("/{id}", handlers.NewRedirectToOriginalURL(urlService))
	router.Get("/ping", handlers.PingDBInit(db))
	router.Get("/api/user/urls", handlers.GetUserURLsHandler(urlService))
	router.Delete("/api/user/urls", handlers.NewBatchDeleteHandler(urlService, workerPool))
	// Подключаем Swagger UI
	router.Get("/swagger/*", httpSwagger.WrapHandler)
	// Запуск HTTP-сервера
	if err := http.ListenAndServe(cfg.ServerAddress, router); err != nil {
		sugar.Errorf("Server error: %v", err)
	}

	return nil
}
