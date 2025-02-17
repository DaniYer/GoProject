package main

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var sugar zap.SugaredLogger

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}

	URL struct {
		URL string `json:"url"`
	}
)

var storage = map[string]string{}

func main() {
	cfg := config.NewConfig()

	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar = *logger.Sugar()

	loadStorageFromFile(cfg.FileStoragePath)

	r := chi.NewRouter()

	r.Post("/", gzipHandle(WithLogging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shortenedURL(w, r, cfg.BaseURL, cfg.FileStoragePath)
	}))))

	r.Get("/{id}", gzipHandle(WithLogging(http.HandlerFunc(redirectedURL))))

	r.Post("/api/shorten", gzipHandle(WithLogging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiShortenHandler(w, r, cfg.BaseURL, cfg.FileStoragePath)
	}))))

	sugar.Infof("Сервер запущен на %s", cfg.ServerAddress)
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		panic(err)
	}
}

func shortenedURL(w http.ResponseWriter, r *http.Request, baseURL, filePath string) {
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	url := string(body)
	genID := genSym()
	storage[genID] = url

	saveStorageToFile(filePath)

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(baseURL + "/" + genID))
}

func redirectedURL(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/")
	if id == "" {
		http.Error(w, "Empty path or not found", http.StatusBadRequest)
		return
	}

	value, exists := storage[id]
	if !exists {
		http.Error(w, "URL not found", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, value, http.StatusTemporaryRedirect)
	w.Header().Set("Content-Type", "text/plain")
}

func apiShortenHandler(w http.ResponseWriter, r *http.Request, baseURL, filePath string) {
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	genID := genSym()
	storage[genID] = req.URL

	saveStorageToFile(filePath)

	resp := struct {
		Result string `json:"result"`
	}{
		Result: baseURL + "/" + genID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func genSym() string {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 8)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

func loadStorageFromFile(filePath string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		sugar.Warnf("Не удалось загрузить данные из файла: %v", err)
		return
	}

	var urls []struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}
	if err := json.Unmarshal(data, &urls); err != nil {
		sugar.Warnf("Ошибка при парсинге JSON: %v", err)
		return
	}

	for _, u := range urls {
		storage[u.ShortURL] = u.OriginalURL
	}
	sugar.Infof("Загружено %d URL из файла", len(storage))
}

func saveStorageToFile(filePath string) {
	var urls []struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}

	for short, original := range storage {
		urls = append(urls, struct {
			ShortURL    string `json:"short_url"`
			OriginalURL string `json:"original_url"`
		}{ShortURL: short, OriginalURL: original})
	}

	data, err := json.Marshal(urls)
	if err != nil {
		sugar.Errorf("Ошибка сериализации данных: %v", err)
		return
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		sugar.Errorf("Ошибка записи в файл: %v", err)
	}
}

func (lw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lw.ResponseWriter.Write(b)
	lw.responseData.size += size
	return size, err
}

func (lw *loggingResponseWriter) WriteHeader(statusCode int) {
	lw.responseData.status = statusCode
	lw.ResponseWriter.WriteHeader(statusCode)
}

func WithLogging(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lw := &loggingResponseWriter{
			ResponseWriter: w,
			responseData:   &responseData{},
		}

		h.ServeHTTP(lw, r)

		duration := time.Since(start)

		sugar.Infow("Request completed",
			"uri", r.RequestURI,
			"method", r.Method,
			"status", lw.responseData.status,
			"duration", duration,
			"size", lw.responseData.size,
		)
	}
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func gzipHandle(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to decode gzip", http.StatusBadRequest)
				return
			}
			defer gzr.Close()
			r.Body = gzr
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			http.Error(w, "Failed to create gzip writer", http.StatusInternalServerError)
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	}
}
