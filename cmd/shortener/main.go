package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	"github.com/DaniYer/GoProject.git/internal/app/handler"
	"github.com/DaniYer/GoProject.git/internal/app/logger"
	"github.com/DaniYer/GoProject.git/internal/app/middleware"
	"github.com/DaniYer/GoProject.git/internal/app/repository"
	"github.com/DaniYer/GoProject.git/internal/app/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.Load()
	logger.Init()
	defer logger.Sync()

	var repo repository.Repository
	var pingFunc func(ctx context.Context) error

	// Подключение к базе
	if cfg.DatabaseDSN != "" {
		pg, err := repository.NewPostgres(cfg.DatabaseDSN)
		if err != nil {
			logger.Log.Fatalf("cannot connect to database: %v", err)
		}
		defer pg.Close(context.Background())
		repo = pg
		pingFunc = pg.Ping
		logger.Log.Infof("Using Postgres storage")
	} else if cfg.FileStoragePath != "" {
		fs, err := repository.NewFileStorage(cfg.FileStoragePath)
		if err != nil {
			logger.Log.Fatalf("cannot open file storage: %v", err)
		}
		repo = fs
		pingFunc = func(ctx context.Context) error { return nil }
		logger.Log.Infof("Using FileStorage: %s", cfg.FileStoragePath)
	} else {
		repo = repository.NewInMemoryRepository()
		pingFunc = func(ctx context.Context) error { return nil }
		logger.Log.Infof("Using In-Memory storage")
	}

	svc := service.NewService(repo)
	h := handler.NewHandler(svc, cfg, pingFunc)

	r := chi.NewRouter()

	r.Use(middleware.GzipDecompressMiddleware)
	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.LoggingMiddleware)

	r.Get("/ping", h.PingHandler)
	r.Post("/", h.ShortenHandler)
	r.Post("/api/shorten", h.ShortenJSONHandler)
	r.Get("/{id}", h.RedirectHandler)
	r.Post("/api/shorten/batch", h.BatchShortenHandler)

	r.Handle("/api/user/urls", middleware.AuthMiddleware(h.UserURLsHandler, cfg.SecretKey))

	server := &http.Server{
		Addr:         cfg.ServerAddress,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	go func() {
		logger.Log.Infof("Starting server on %s", cfg.ServerAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatalf("Server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Log.Errorf("Shutdown error: %v", err)
	}

	logger.Log.Infof("Server stopped gracefully")
}
