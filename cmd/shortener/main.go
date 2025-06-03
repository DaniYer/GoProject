package main

import (
	"database/sql"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/batch"
	"github.com/DaniYer/GoProject.git/internal/app/config"
	"github.com/DaniYer/GoProject.git/internal/app/gzipmiddleware"
	"github.com/DaniYer/GoProject.git/internal/app/logging"
	"github.com/DaniYer/GoProject.git/internal/app/ping"
	"github.com/DaniYer/GoProject.git/internal/app/redirect"
	"github.com/DaniYer/GoProject.git/internal/app/shortener"
	"github.com/DaniYer/GoProject.git/internal/app/storage/database"
	"github.com/DaniYer/GoProject.git/internal/app/storage/file"
	"github.com/DaniYer/GoProject.git/internal/app/storage/memory"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func main() {
	// Инициализация конфигурации
	cfg := config.NewConfig()

	// Логгер
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()
	logging.InitLogger(sugar)

	var (
		db          *sql.DB
		store       shortener.URLStore
		storeWithDB shortener.URLStoreWithDBforHandler
	)

	// Выбор хранилища (БД → файл → память)
	if cfg.DatabaseDSN != "" && cfg.DatabaseDSN != "localDB" {
		db, err = database.InitDB("pgx", cfg.DatabaseDSN)
		if err != nil {
			sugar.Errorf("Ошибка подключения к БД: %v", err)
		} else {
			if err := database.CreateTable(db); err != nil {
				sugar.Errorf("Ошибка создания таблицы: %v", err)
			} else {
				dbStore := database.NewDBStore(db)
				store = dbStore
				storeWithDB = dbStore
			}
		}
	}

	if store == nil && cfg.FileStoragePath != "" {
		fs, err := file.NewFileStore(cfg.FileStoragePath)
		if err != nil {
			sugar.Errorf("Ошибка инициализации файлового хранилища: %v", err)
		} else {
			store = fs
			storeWithDB = fs
		}
	}

	if store == nil {
		sugar.Infof("Используется in-memory хранилище")
		memStore := memory.NewMemoryStore()
		store = memStore
		storeWithDB = memStore
	}

	// Роутер
	router := chi.NewRouter()

	sugar.Infow("Запуск сервера", "адрес", cfg.ServerAddress)

	router.Use(logging.WithLogging)
	router.Use(gzipmiddleware.GzipHandle)

	// Роуты
	router.Post("/", shortener.NewGenerateShortURLHandler(cfg, storeWithDB, sugar))
	router.Post("/api/shorten", shortener.NewHandleShortenURLv13(cfg, storeWithDB, sugar))
	router.Post("/api/shorten/batch", batch.NewBatchShortenURLHandler(cfg.BaseURL, store))
	router.Get("/{id}", redirect.NewRedirectToOriginalURL(store))
	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) { ping.PingDB(db, w) })

	if err := http.ListenAndServe(cfg.ServerAddress, router); err != nil {
		sugar.Errorf("Ошибка сервера: %v", err)
	}
}
