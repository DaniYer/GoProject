package main

import (
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	"github.com/DaniYer/GoProject.git/internal/app/database"
	gziphandle "github.com/DaniYer/GoProject.git/internal/app/gzipmiddleware"
	"github.com/DaniYer/GoProject.git/internal/app/logging"
	"github.com/DaniYer/GoProject.git/internal/app/redirect"
	"github.com/DaniYer/GoProject.git/internal/app/shortener"
	"github.com/DaniYer/GoProject.git/internal/app/storage"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"go.uber.org/zap"
)

var (
	sugar *zap.SugaredLogger
	cfg   = config.ConfigInit()
)

func main() {

	router := chi.NewRouter()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// Создаём SugaredLogger и передаём его в logging.InitLogger()
	sugar = logger.Sugar()
	logging.InitLogger(sugar)

	sugar.Infow(
		"Starting server",
		"addr", cfg.A,
	)
	write, _ := storage.NewFileStorage(cfg.F)
	read, _ := storage.NewConsumer(cfg.F)

	router.Use(logging.WithLogging) // Теперь логгер в logging будет работать
	router.Use(gziphandle.GzipHandle)
	router.Post("/", shortener.NewGenerateShortURLHandler(cfg, write))
	router.Get("/{id}", redirect.NewRedirectToOriginalURL(read))
	router.Post("/api/shorten", shortener.NewHandleShortenURL(cfg, write))
	router.Get("/ping", database.Connect(cfg))

	if err := http.ListenAndServe(cfg.A, router); err != nil {
		sugar.Errorf("RIP %v", err) // исправлено форматирование ошибки
	}
}
