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
		return fmt.Errorf("init logger err: %w", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()
	middlewares.InitLogger(sugar)

	var (
		db    *sql.DB
		store service.URLStore
	)

	if cfg.DatabaseDSN != "" {
		db, err = database.InitDB("pgx", cfg.DatabaseDSN)
		if err != nil {
			sugar.Errorf("DB connect error: %v", err)
			return err
		}
		if err := goose.Up(db, "internal/app/storage/database/migrations"); err != nil {
			sugar.Errorf("Migrations error: %v", err)
			return err
		}
		store = database.NewDBStore(db)
	}

	if store == nil && cfg.FileStoragePath != "" {
		fs, err := file.NewFileStore(cfg.FileStoragePath)
		if err != nil {
			sugar.Errorf("File store error: %v", err)
		} else {
			store = fs
		}
	}

	if store == nil {
		sugar.Infof("Fallback in-memory store")
		store = memory.NewMemoryStore()
	}

	urlService := service.URLService{
		Store:   store,
		BaseURL: cfg.BaseURL,
	}

	// создаем worker pool
	pool := worker.NewDeleteWorkerPool(store, 1024, 10, 2*time.Second)
	pool.Start()
	defer pool.Shutdown()

	router := chi.NewRouter()

	sugar.Infow("Start server", "addr", cfg.ServerAddress)

	router.Use(middlewares.WithLogging)
	router.Use(middlewares.GzipHandle)
	router.Use(middlewares.AuthMiddleware)

	router.Post("/", handlers.NewGenerateShortURLHandler(&urlService))
	router.Post("/api/shorten/batch", handlers.NewBatchShortenURLHandler(&urlService))
	router.Post("/api/shorten", handlers.NewHandleShortenURLv13(&urlService))
	router.Get("/{id}", handlers.NewRedirectToOriginalURL(&urlService))
	router.Get("/ping", handlers.PingDBInit(db))
	router.Get("/api/user/urls", handlers.GetUserURLsHandler(&urlService))
	router.Delete("/api/user/urls", (&handlers.DeleteHandler{
		Svc:  &urlService,
		Pool: pool,
	}).ServeHTTP)

	if err := http.ListenAndServe(cfg.ServerAddress, router); err != nil {
		sugar.Errorf("Server error: %v", err)
	}
	return nil
}
