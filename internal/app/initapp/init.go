package initapp

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/batch"
	"github.com/DaniYer/GoProject.git/internal/app/config"
	"github.com/DaniYer/GoProject.git/internal/app/gzipmiddleware"
	"github.com/DaniYer/GoProject.git/internal/app/logging"
	"github.com/DaniYer/GoProject.git/internal/app/ping"
	"github.com/DaniYer/GoProject.git/internal/app/redirect"
	"github.com/DaniYer/GoProject.git/internal/app/service"
	"github.com/DaniYer/GoProject.git/internal/app/shortener"
	"github.com/DaniYer/GoProject.git/internal/app/storage/database"
	"github.com/DaniYer/GoProject.git/internal/app/storage/file"
	"github.com/DaniYer/GoProject.git/internal/app/storage/memory"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"
	"go.uber.org/zap"
)

func InitializeApp() error {
	cfg := config.NewConfig()

	// Логгер
	logger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("ошибка инициализации логгера: %w", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()
	logging.InitLogger(sugar)

	// Выбираем хранилище
	var (
		db    *sql.DB
		store service.URLStore
	)

	if cfg.DatabaseDSN != "" && cfg.DatabaseDSN != config.DefaultDatabaseDSN {
		db, err = database.InitDB("pgx", cfg.DatabaseDSN)
		if err != nil {
			sugar.Errorf("Ошибка подключения к БД: %v", err)
			return err
		}

		if err := goose.Up(db, "internal/app/storage/database/migrations"); err != nil {
			sugar.Errorf("Ошибка применения миграций: %v", err)
			return err
		}

		store = database.NewDBStore(db)
	}

	if store == nil && cfg.FileStoragePath != "" {
		fs, err := file.NewFileStore(cfg.FileStoragePath, sugar)
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

	urlService := service.URLService{
		Store:   store,
		BaseURL: cfg.BaseURL,
	}

	router := chi.NewRouter()

	sugar.Infow("Запуск сервера", "адрес", cfg.ServerAddress)

	router.Use(logging.WithLogging)
	router.Use(gzipmiddleware.GzipHandle)

	// Роуты
	router.Post("/", shortener.NewGenerateShortURLHandler(&urlService))
	router.Post("/api/shorten/batch", batch.NewBatchShortenURLHandler(&urlService))
	router.Get("/{id}", redirect.NewRedirectToOriginalURL(store))
	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) { ping.PingDB(db, w) })

	router.Post("/api/shorten", shortener.NewHandleShortenURLv13(&urlService))

	if err := http.ListenAndServe(cfg.ServerAddress, router); err != nil {
		sugar.Errorf("Ошибка сервера: %v", err)
	}
	return nil
}
