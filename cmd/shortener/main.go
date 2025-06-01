package main

import (
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	gziphandle "github.com/DaniYer/GoProject.git/internal/app/gzipmiddleware"
	"github.com/DaniYer/GoProject.git/internal/app/logging"
	"github.com/DaniYer/GoProject.git/internal/app/redirect"
	"github.com/DaniYer/GoProject.git/internal/app/shortener"
	"github.com/DaniYer/GoProject.git/internal/app/storage"
	"github.com/go-chi/chi/v5"

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
	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		shortener.GenerateShortURLHandler(w, r, cfg, write)
	})
	router.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		redirect.RedirectToOriginalURL(w, r, read)
	})
	router.Post("/api/shorten", func(w http.ResponseWriter, r *http.Request) {
		shortener.HandleShortenURL(w, r, cfg, write)
	})

	if err := http.ListenAndServe(cfg.A, router); err != nil {
		sugar.Errorf("RIP %v", err) // исправлено форматирование ошибки
	}
}
