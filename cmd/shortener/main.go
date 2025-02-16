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
		http.ResponseWriter
		responseData     *responseData
		responseDataBody string
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

func WithLogging(h http.HandlerFunc) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}

		h.ServeHTTP(&lw, r) // внедряем реализацию http.ResponseWriter

		duration := time.Since(start)

		sugar.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status, // получаем перехваченный код статуса ответа
			"duration", duration,
			"size", responseData.size, // получаем перехваченный размер ответа
		)
	}
	return logFn
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

func (lw *loggingResponseWriter) Write(b []byte) (int, error) {
	lw.responseDataBody = string(b) // Сохраняем тело
	return len(b), nil
}

func (lw *loggingResponseWriter) WriteHeader(statusCode int) {
	lw.responseData.status = statusCode
	lw.ResponseWriter.WriteHeader(statusCode)
}

func gzipHandle(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, поддерживает ли клиент gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Создаем gzip.Writer поверх оригинального ResponseWriter
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			http.Error(w, "Failed to create gzip writer", http.StatusInternalServerError)
			return
		}
		defer gz.Close()

		// Обертка для перехвата заголовков и статуса
		lw := &loggingResponseWriter{
			ResponseWriter: w,
			responseData:   &responseData{},
		}

		next.ServeHTTP(lw, r)

		// Условие сжатия только для JSON и HTML
		contentType := lw.Header().Get("Content-Type")
		if lw.responseData.status < 300 && (strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html")) {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Del("Content-Length")
			_, _ = gz.Write([]byte(lw.responseDataBody))
		} else {
			// Если контент не подлежит сжатию, отдаём как есть
			w.WriteHeader(lw.responseData.status)
			_, _ = w.Write([]byte(lw.responseDataBody))
		}
	}
}
