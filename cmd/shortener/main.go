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
	"strings"
	"time"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

var sugar *zap.SugaredLogger

// Record представляет запись для хранения URL.
type Record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// URLStore описывает методы для сохранения и получения URL.
type URLStore interface {
	Save(shortURL, originalURL string) error
	Get(shortURL string) (string, error)
}

// DBStore – реализация URLStore через PostgreSQL.
type DBStore struct {
	db *sql.DB
}

func NewDBStore(db *sql.DB) *DBStore {
	return &DBStore{db: db}
}

func createTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		short_url VARCHAR(8) UNIQUE NOT NULL,
		original_url TEXT NOT NULL
	);`
	_, err := db.Exec(query)
	return err
}

func (s *DBStore) Save(shortURL, originalURL string) error {
	query := "INSERT INTO urls (short_url, original_url) VALUES ($1, $2)"
	_, err := s.db.Exec(query, shortURL, originalURL)
	return err
}

func (s *DBStore) Get(shortURL string) (string, error) {
	var originalURL string
	query := "SELECT original_url FROM urls WHERE short_url=$1"
	err := s.db.QueryRow(query, shortURL).Scan(&originalURL)
	if err != nil {
		return "", err
	}
	return originalURL, nil
}

// MemoryStore – реализация URLStore с использованием in-memory карты.
type MemoryStore struct {
	data map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[string]string)}
}

func (ms *MemoryStore) Save(shortURL, originalURL string) error {
	ms.data[shortURL] = originalURL
	return nil
}

func (ms *MemoryStore) Get(shortURL string) (string, error) {
	if url, ok := ms.data[shortURL]; ok {
		return url, nil
	}
	return "", fmt.Errorf("not found")
}

// FileStore – реализация URLStore, использующая файл для персистентности.
type FileStore struct {
	filePath string
	data     map[string]string
}

func NewFileStore(filePath string) (*FileStore, error) {
	store, err := loadStorageFromFile(filePath)
	if err != nil {
		return nil, err
	}
	return &FileStore{filePath: filePath, data: store}, nil
}

func (fs *FileStore) Save(shortURL, originalURL string) error {
	fs.data[shortURL] = originalURL
	rec := Record{
		UUID:        "", // для FileStore не требуется генерация UUID
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
	return appendRecordToFile(rec, fs.filePath)
}

func (fs *FileStore) Get(shortURL string) (string, error) {
	if url, ok := fs.data[shortURL]; ok {
		return url, nil
	}
	return "", fmt.Errorf("not found")
}

// loadStorageFromFile загружает записи из файла в карту.
func loadStorageFromFile(filePath string) (map[string]string, error) {
	store := make(map[string]string)
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return store, nil
		}
		return nil, err
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
		store[rec.ShortURL] = rec.OriginalURL
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return store, nil
}

// appendRecordToFile дописывает запись в файл.
func appendRecordToFile(rec Record, filePath string) error {
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

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	sugar = logger.Sugar()

	cfg := config.NewConfig()
	sugar.Infof("Configuration: ServerAddress=%s, BaseURL=%s, FileStoragePath=%s, DatabaseDSN=%s",
		cfg.ServerAddress, cfg.BaseURL, cfg.FileStoragePath, cfg.DatabaseDSN)

	var store URLStore

	// Если задана корректная строка подключения к PostgreSQL, используем DBStore.
	if cfg.DatabaseDSN != "" && cfg.DatabaseDSN != "localDB" {
		db, err := initDB("postgres", cfg.DatabaseDSN)
		if err != nil {
			sugar.Errorf("Ошибка подключения к БД: %v", err)
		} else {
			if err := createTable(db); err != nil {
				sugar.Errorf("Ошибка создания таблицы: %v", err)
			} else {
				store = NewDBStore(db)
			}
		}
	}
	// Если DBStore не инициализирована, пытаемся файловое хранилище.
	if store == nil && cfg.FileStoragePath != "" {
		fs, err := NewFileStore(cfg.FileStoragePath)
		if err != nil {
			sugar.Errorf("Ошибка инициализации файлового хранилища: %v", err)
		} else {
			store = fs
		}
	}
	// Если ни один из вариантов не сработал – используем in-memory хранилище.
	if store == nil {
		sugar.Infof("Используется in-memory хранилище")
		store = NewMemoryStore()
	}

	r := chi.NewRouter()
	r.Use(gzipMiddleware)
	r.Post("/", WithLogging(func(w http.ResponseWriter, r *http.Request) {
		shortenedURL(w, r, cfg.BaseURL, store)
	}))
	r.Get("/{id}", WithLogging(func(w http.ResponseWriter, r *http.Request) {
		redirectedURL(w, r, store)
	}))
	r.Post("/api/shorten", WithLogging(func(w http.ResponseWriter, r *http.Request) {
		jsonHandler(w, r, store)
	}))
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		pingHandler(w, r, cfg.DatabaseDSN)
	})
	r.Post("/api/shorten/batch", func(w http.ResponseWriter, r *http.Request) {
		batchShortenHandler(w, r, "http://localhost:8080", NewMemoryStore())
	})

	fmt.Println("Сервер запущен на", cfg.ServerAddress)
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		panic(err)
	}
}

// initDB устанавливает соединение с базой данных.
func initDB(driverName, dataSourceName string) (*sql.DB, error) {
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

// shortenedURL обрабатывает POST-запрос для создания короткого URL.
func shortenedURL(w http.ResponseWriter, r *http.Request, baseURL string, store URLStore) {
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
	originalURL := string(body)

	shortURL := genSym()

	if err := store.Save(shortURL, originalURL); err != nil {
		sugar.Errorf("Ошибка сохранения URL: %v", err)
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(baseURL + "/" + shortURL))
}

// redirectedURL обрабатывает GET-запрос для редиректа по короткому URL.
func redirectedURL(w http.ResponseWriter, r *http.Request, store URLStore) {
	if r.Method != http.MethodGet {
		http.Error(w, "Unresolved method", 400)
		return
	}
	id := r.URL.Path[1:]
	if id == "" {
		http.Error(w, "Empty path or not found", 400)
		return
	}
	originalURL, err := store.Get(id)
	if err != nil || originalURL == "" {
		http.Error(w, "URL not found", 400)
		return
	}
	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

// jsonHandler обрабатывает JSON API для создания короткого URL.
func jsonHandler(w http.ResponseWriter, r *http.Request, store URLStore) {
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
	if err := store.Save(shortURL, req.URL); err != nil {
		sugar.Errorf("Ошибка сохранения в хранилище: %v", err)
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
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

// genSym генерирует строку из 8 случайных строчных букв.
func genSym() string {
	result := ""
	for i := 0; i < 8; i++ {
		result += string(rune(rand.Intn(26) + 'a'))
	}
	return result
}

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

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

// iter12
type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func batchShortenHandler(w http.ResponseWriter, r *http.Request, baseURL string, store URLStore) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var requests []BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(requests) == 0 {
		http.Error(w, "Empty batch request", http.StatusBadRequest)
		return
	}

	responses := make([]BatchResponse, len(requests))

	for i, req := range requests {
		shortURL := genSym()
		if err := store.Save(shortURL, req.OriginalURL); err != nil {
			http.Error(w, "Error saving URL", http.StatusInternalServerError)
			return
		}
		responses[i] = BatchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      baseURL + "/" + shortURL,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responses)
}
