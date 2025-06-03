package main

import (
	"database/sql"
	"net/http"

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
	sugar *zap.SugaredLogger
	cfg   = config.NewConfig()
	db    *sql.DB
)

type URLStore interface {
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
				store = database.NewDBStore(db)
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

	router.Post("/", shortener.NewGenerateShortURLHandler(cfg, store))
	router.Get("/{id}", redirect.NewRedirectToOriginalURL(store))
	router.Post("/api/shorten", shortener.NewHandleShortenURL(cfg, store))
	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		ping.PingDB(db, w)
	})

	if err := http.ListenAndServe(cfg.ServerAddress, router); err != nil {
		sugar.Errorf("RIP %v", err)
	}
}

// This is the main entry point for the URL shortener service.
// It initializes the configuration, sets up the database or file storage,
// configures the HTTP router with the necessary handlers, and starts the HTTP server.
