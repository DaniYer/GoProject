package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var sugar *zap.SugaredLogger
var storage *Storage

func main() {
	cfg := config.NewConfig()
	logger, _ := zap.NewDevelopment()
	sugar = logger.Sugar()
	defer logger.Sync()

	// Создаём файловое хранилище
	var err error
	storage, err = NewStorage(cfg.FileStoragePath)
	if err != nil {
		panic("Ошибка загрузки хранилища: " + err.Error())
	}

	// Подключаем маршруты
	r := chi.NewRouter()
	r.Use(gzipMiddleware)
	r.Post("/", WithLogging(func(w http.ResponseWriter, r *http.Request) {
		shortenedURL(w, r, cfg.BaseURL)
	}))
	r.Get("/{id}", WithLogging(redirectedURL))
	r.Post("/api/shorten", jsonHandler)

	fmt.Println("Запуск сервера на", cfg.ServerAddress)

	if err = http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		sugar.Fatalf("Ошибка запуска сервера: %v", err)
	}

}

// shortenedURL принимает URL, сохраняет его и возвращает сокращённую ссылку
func shortenedURL(w http.ResponseWriter, r *http.Request, baseURL string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Unresolved method", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	url := string(body)
	genID := genSym()

	// Сохраняем в файл
	if err := storage.SaveURL(genID, url); err != nil {
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(baseURL + "/" + genID))
}

// redirectedURL принимает сокращенный URL и редиректит на оригинальный
func redirectedURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Empty path or not found", http.StatusBadRequest)
		return
	}

	value, exists := storage.GetURL(id)
	if !exists {
		http.Error(w, "URL not found", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, value, http.StatusTemporaryRedirect)
}

// jsonHandler обрабатывает JSON-запрос с URL
func jsonHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.URL == "" {
		http.Error(w, "Empty URL", http.StatusBadRequest)
		return
	}

	genID := genSym()

	// Сохраняем в файл
	if err := storage.SaveURL(genID, req.URL); err != nil {
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	resp := struct {
		Result string `json:"result"`
	}{
		Result: fmt.Sprintf("%s/%s", r.Host, genID),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// WithLogging — middleware для логирования запросов
func WithLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sugar.Infof("Запрос: %s %s", r.Method, r.URL.Path)
		next(w, r)
	}
}

// gzipMiddleware — middleware для работы с gzip

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, сжат ли запрос
		if r.Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to read gzip body", http.StatusInternalServerError)
				return
			}
			defer reader.Close()
			r.Body = reader
		}

		// Проверяем, поддерживает ли клиент GZIP
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Создаём gzip.Writer для ответа
		gz := gzip.NewWriter(w)
		defer gz.Close()

		// Устанавливаем заголовок только если клиент поддерживает gzip
		w.Header().Set("Content-Encoding", "gzip")

		gzipWriter := &gzipResponseWriter{Writer: gz, ResponseWriter: w}
		next.ServeHTTP(gzipWriter, r)
	})
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// genSym генерирует короткий идентификатор
func genSym() string {
	return fmt.Sprintf("%06d", 100000+os.Getpid()%900000) // Простой генератор
}
