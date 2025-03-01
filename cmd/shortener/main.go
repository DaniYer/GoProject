package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

var storage = map[string]string{}
var sugar *zap.SugaredLogger

// Глобальные переменные для работы с персистентным хранилищем
var nextUUID int = 1
var storageFilePath string

// Record представляет запись, которая сохраняется в файл
type Record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func main() {
	// создаём предустановленный регистратор zap
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// Получаем конфигурацию из пакета config
	cfg := config.NewConfig()
	sugar = logger.Sugar()

	// Сохраняем путь до файла в глобальной переменной
	storageFilePath = cfg.FileStoragePath

	// Выводим конфигурацию (для отладки)
	cfg.Print()

	// Загружаем ранее сохранённые URL из файла, если он существует
	if err := loadStorage(storageFilePath); err != nil {
		sugar.Errorf("Ошибка загрузки данных из файла: %v", err)
	}

	r := chi.NewRouter()
	r.Use(gzipMiddleware)
	r.Post("/", WithLogging(func(w http.ResponseWriter, r *http.Request) {
		shortenedURL(w, r, cfg.BaseURL)
	}))
	r.Get("/{id}", WithLogging(redirectedURL))
	r.Post("/api/shorten", WithLogging(jsonHandler))
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		pingHandler(w, r, cfg.DatabaseDSN)
	})

	fmt.Println(cfg.ServerAddress)
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		panic(err)
	}
}

// loadStorage загружает данные из файла в память.
func loadStorage(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		// Если файла нет, ничего не загружаем
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var rec Record
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			sugar.Errorf("Ошибка парсинга записи: %v", err)
			continue
		}
		storage[rec.ShortURL] = rec.OriginalURL
		// Обновляем nextUUID, если найденное значение UUID больше текущего
		if id, err := strconv.Atoi(rec.UUID); err == nil && id >= nextUUID {
			nextUUID = id + 1
		}
	}
	return scanner.Err()
}

// appendRecord дописывает запись в файл.
func appendRecord(rec Record, filePath string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	recBytes, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	_, err = file.Write(append(recBytes, '\n'))
	return err
}

// shortenedURL обрабатывает POST-запрос для создания короткого URL.
func shortenedURL(w http.ResponseWriter, r *http.Request, baseURL string) {
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

	// Генерация короткого URL
	shortURL := genSym()
	storage[shortURL] = url

	record := Record{
		UUID:        strconv.Itoa(nextUUID),
		ShortURL:    shortURL,
		OriginalURL: url,
	}
	nextUUID++
	if err := appendRecord(record, storageFilePath); err != nil {
		sugar.Errorf("Ошибка записи в файл: %v", err)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(201)
	w.Write([]byte(baseURL + "/" + shortURL))
}

// redirectedURL обрабатывает GET-запрос для редиректа.
func redirectedURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Unresolved method", 400)
		return
	}
	id := r.URL.Path[1:]
	if id == "" {
		http.Error(w, "Empty path or not found", 400)
		return
	}
	value, exists := storage[id]
	if !exists {
		http.Error(w, "URL not found", 400)
		return
	}
	http.Redirect(w, r, value, http.StatusTemporaryRedirect)
	w.Header().Set("Content-Type", "text/plain")
}

// genSym генерирует строку из 8 случайных строчных букв.
func genSym() string {
	result := ""
	for i := 0; i < 8; i++ {
		result += string(rune(rand.Intn(26) + 'a'))
	}
	return result
}

type (
	// responseData хранит сведения об ответе для логирования.
	responseData struct {
		status int
		size   int
	}

	// loggingResponseWriter расширяет http.ResponseWriter для логирования.
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func WithLogging(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h(&lw, r)

		duration := time.Since(start)

		sugar.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}
}

// jsonHandler обрабатывает JSON-запрос для создания короткого URL.
func jsonHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения тела запроса", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req shortenURL
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Ошибка парсинга JSON", http.StatusBadRequest)
		return
	}

	shortURL := genSym()
	storage[shortURL] = req.URL

	record := Record{
		UUID:        strconv.Itoa(nextUUID),
		ShortURL:    shortURL,
		OriginalURL: req.URL,
	}
	nextUUID++
	if err := appendRecord(record, storageFilePath); err != nil {
		sugar.Errorf("Ошибка записи в файл: %v", err)
	}

	resp := redirectURL{
		Result: "http://localhost:8080/" + shortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

type shortenURL struct {
	URL string `json:"url"`
}

type redirectURL struct {
	Result string `json:"result"`
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Ошибка распаковки GZIP", http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = io.NopCloser(gz)
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gzWriter := gzip.NewWriter(w)
		defer gzWriter.Close()

		gzResponse := gzipResponseWriter{ResponseWriter: w, Writer: gzWriter}
		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(&gzResponse, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// initDB устанавливает соединение с базой данных и выполняет проверку (ping).
func initDB(driverName string, dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// pingHandler проверяет соединение с базой данных и возвращает 200 OK при успехе,
// либо 500 Internal Server Error, если подключение не удалось.
func pingHandler(w http.ResponseWriter, r *http.Request, dataSN string) {
	db, err := initDB("postgres", dataSN)
	if err != nil {
		sugar.Errorf("Ошибка чтения базы данных: %v", err)
		http.Error(w, "Ошибка соединения с БД", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Если соединение установлено, возвращаем сообщение (по умолчанию статус 200 OK)
	w.Write([]byte("Связь налажена"))
}
