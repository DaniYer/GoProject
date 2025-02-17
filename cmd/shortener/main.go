package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var sugar zap.SugaredLogger

// должно работать
type (
	// берём структуру для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
	URL struct {
		URL string `json:"url"`
	}
)

var storage = map[string]string{}

func main() {

	// Получаем конфигурацию из пакета config
	cfg := config.NewConfig()

	logger, _ := zap.NewProduction()
	defer logger.Sync() // Обязательно синхронизировать перед выходом
	sugar = *logger.Sugar()

	// Выводим конфигурацию (для отладки)
	cfg.Print()
	r := chi.NewRouter()
	r.Post("/", gzipHandle(WithLogging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shortenedURL(w, r, cfg.BaseURL)
	}))))
	r.Get("/{id}", gzipHandle(WithLogging(http.HandlerFunc(redirectedURL))))

	r.Post("/api/shorten", gzipHandle(WithLogging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiShortenHandler(w, r, cfg.BaseURL)
	}))))

	fmt.Println(cfg.ServerAddress)
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		panic(err)
	}
}
func shortenedURL(w http.ResponseWriter, r *http.Request, cfg string) {
	// Только POST запросы
	if r.Method != http.MethodPost {
		http.Error(w, "Unresolved method", 400)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	url := string(body)

	// Генерация уникального идентификатора
	genID := genSym()

	// Сохранение в мапу
	storage[genID] = url

	// Отправка ответа с сокращенным URL
	w.WriteHeader(201)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(cfg + "/" + genID))
}

func redirectedURL(w http.ResponseWriter, r *http.Request) {
	// Только GET запросы
	if r.Method != http.MethodGet {
		http.Error(w, "Unresolved method", 400)
		return
	}
	// Извлечение идентификатора из пути
	id := r.URL.Path[1:]
	if id == "" {
		http.Error(w, "Empty path or not found", 400)
		return
	}
	// Поиск в мапе
	value, exists := storage[id]
	if !exists {
		http.Error(w, "URL not found", 400)
		return
	}
	// Редирект на оригинальный URL
	http.Redirect(w, r, value, http.StatusTemporaryRedirect)
	w.Header().Set("Content-Type", "text/plain")
}
func genSym() string {
	result := ""
	for i := 0; i < 8; i++ {
		result += string(rune(rand.Intn(26) + 'a')) // Генерация случайных символов
	}
	return result
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

		// Используем кастомный ResponseWriter для логирования
		lw := &loggingResponseWriter{
			ResponseWriter: w,
			responseData:   &responseData{},
		}

		h.ServeHTTP(lw, r)

		duration := time.Since(start)

		// Логируем данные
		sugar.Infow("Request completed",
			"uri", r.RequestURI,
			"method", r.Method,
			"status", lw.responseData.status,
			"duration", duration,
			"size", lw.responseData.size,
		)
	}
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Result string `json:"result"`
}

func apiShortenHandler(w http.ResponseWriter, r *http.Request, baseURL string) {
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	genID := genSym()
	storage[genID] = req.URL

	resp := shortenResponse{
		Result: baseURL + "/" + genID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

func gzipHandle(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 1. Распаковка входного запроса, если он сжат
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to decode gzip", http.StatusBadRequest)
				return
			}
			defer gzr.Close()
			r.Body = gzr
		}

		// 2. Проверка, поддерживает ли клиент сжатые ответы
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// 3. Создание gzip.Writer поверх текущего ResponseWriter
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			http.Error(w, "Failed to create gzip writer", http.StatusInternalServerError)
			return
		}
		defer gz.Close()

		// Устанавливаем заголовки для сжатого ответа
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		// Передаём управление дальше с обёрнутым gzipWriter
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	}
}
