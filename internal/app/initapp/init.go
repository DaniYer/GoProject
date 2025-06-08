package initapp

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

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
		return fmt.Errorf("init logger error: %w", err)
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
			sugar.Errorf("DB connect error: %v", err)
			return err
		}
		if err := goose.Up(db, "internal/app/storage/database/migrations"); err != nil {
			sugar.Errorf("Migration error: %v", err)
			return err
		}
		store = database.NewDBStore(db)
	}

	if store == nil && cfg.FileStoragePath != "" {
		fs, err := file.NewFileStore(cfg.FileStoragePath)
		if err != nil {
			sugar.Errorf("FileStore init error: %v", err)
		} else {
			store = fs
		}
	}

	if store == nil {
		sugar.Infof("Using in-memory storage")
		store = memory.NewMemoryStore()
	}

	urlService := service.NewURLService(store, cfg.BaseURL)

	// запускаем worker pool
	workerPool := worker.NewDeleteWorkerPool(urlService, 1024, 5*time.Second)
	workerPool.Start()

	router := chi.NewRouter()

	sugar.Infow("Start server", "addr", cfg.ServerAddress)

	router.Use(middlewares.WithLogging)
	router.Use(middlewares.GzipHandle)
	router.Use(middlewares.AuthMiddleware)

	router.Post("/", handlers.NewGenerateShortURLHandler(urlService))
	router.Post("/api/shorten", handlers.NewHandleShortenURLv13(urlService))
	router.Post("/api/shorten/batch", handlers.NewBatchShortenURLHandler(urlService))
	router.Get("/{id}", handlers.NewRedirectToOriginalURL(urlService))
	router.Get("/ping", handlers.PingDBInit(db))
	router.Get("/api/user/urls", handlers.GetUserURLsHandler(urlService))
	router.Delete("/api/user/urls", handlers.NewBatchDeleteHandler(urlService, workerPool))

	if err := http.ListenAndServe(cfg.ServerAddress, router); err != nil {
		sugar.Errorf("Server error: %v", err)
	}

	return nil
}
