package initapp

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	"github.com/DaniYer/GoProject.git/internal/app/handlers"
	"github.com/DaniYer/GoProject.git/internal/app/middlewares"
	"github.com/DaniYer/GoProject.git/internal/app/service"
	"github.com/DaniYer/GoProject.git/internal/app/storage/database"
	"github.com/DaniYer/GoProject.git/internal/app/storage/file"
	"github.com/DaniYer/GoProject.git/internal/app/storage/memory"
	"github.com/DaniYer/GoProject.git/internal/app/worker"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"
	"go.uber.org/zap"
)

func InitializeApp() error {
	cfg := config.NewConfig()

	logger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("ошибка инициализации логгера: %w", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()
	middlewares.InitLogger(sugar)

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

	urlService := service.URLService{
		Store:   store,
		BaseURL: cfg.BaseURL,
	}

	// создаем worker
	deleteWorker := worker.NewDeleteWorker(urlService.Store, 100)
	deleteWorker.Start()

	router := chi.NewRouter()

	sugar.Infow("Запуск сервера", "адрес", cfg.ServerAddress)

	router.Use(middlewares.WithLogging)
	router.Use(middlewares.GzipHandle)
	router.Use(middlewares.AuthMiddleware)

	// Роуты
	router.Post("/", handlers.NewGenerateShortURLHandler(&urlService))
	router.Post("/api/shorten/batch", handlers.NewBatchShortenURLHandler(&urlService))
	router.Get("/{id}", handlers.NewRedirectToOriginalURL(&urlService))

	router.Get("/ping", handlers.PingDBInit(db))
	router.Post("/api/shorten", handlers.NewHandleShortenURLv13(&urlService))
	router.Get("/api/user/urls", handlers.GetUserURLsHandler(&urlService))
	router.Delete("/api/user/urls", handlers.NewBatchDeleteHandler(handlers.BatchDeleteDeps{
		Service: &urlService,
		Worker:  deleteWorker,
	}))

	if err := http.ListenAndServe(cfg.ServerAddress, router); err != nil {
		sugar.Errorf("Ошибка сервера: %v", err)
	}
	return nil
}
