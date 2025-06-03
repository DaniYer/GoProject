package main

import (
	"database/sql"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/batch"
	"github.com/DaniYer/GoProject.git/internal/app/config"
	gziphandle "github.com/DaniYer/GoProject.git/internal/app/gzipmiddleware"
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

var (
	sugar       *zap.SugaredLogger
	cfg         = config.NewConfig()
	db          *sql.DB
	storeWithDB shortener.URLStoreWithDBforHandler
)

type URLStore interface {
	SaveWithConflict(shortURL, originalURL string) (string, error)
	Save(shortURL, originalURL string) error
	Get(shortURL string) (string, error)
}

func main() {
	var store URLStore

	if cfg.DatabaseDSN != "" && cfg.DatabaseDSN != "localDB" {
		var err error
		db, err = database.InitDB("pgx", cfg.DatabaseDSN)
		if err != nil {
			sugar.Errorf("Ошибка подключения к БД: %v", err)
		} else {
			if err := database.CreateTable(db); err != nil {
				sugar.Errorf("Ошибка создания таблицы: %v", err)
			} else {
				dbStore := database.NewDBStore(db)
				store = dbStore
				storeWithDB = dbStore // он реализует расширенный интерфейс
			}
		}
	}

	if store == nil && cfg.FileStoragePath != "" {
		fs, err := file.NewFileStore(cfg.FileStoragePath)
		if err != nil {
			sugar.Errorf("Ошибка инициализации файлового хранилища: %v", err)
		} else {
			store = fs
		}
	}

	if store == nil {
		sugar.Infof("Используется in-memory хранилище")
		store = memory.NewMemoryStore()
	}

	router := chi.NewRouter()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	sugar = logger.Sugar()
	logging.InitLogger(sugar)

	sugar.Infow("Starting server", "addr", cfg.ServerAddress)

	router.Use(logging.WithLogging)
	router.Use(gziphandle.GzipHandle)

	router.Post("/", shortener.NewGenerateShortURLHandler(cfg, storeWithDB))        // Итерация 1 и 13
	router.Post("/api/shorten", shortener.NewHandleShortenURLv13(cfg, storeWithDB)) // Итерация 7 и 13

	router.Get("/{id}", redirect.NewRedirectToOriginalURL(store))
	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		ping.PingDB(db, w)
	})
	router.Post("/api/shorten/batch", batch.NewBatchShortenURLHandler(cfg.BaseURL, store))

	if err := http.ListenAndServe(cfg.ServerAddress, router); err != nil {
		sugar.Errorf("RIP %v", err)
	}
}
