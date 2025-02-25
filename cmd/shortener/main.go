package main

import (
	"bufio"
	"compress/gzip"
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
	"go.uber.org/zap"
)

// git
var storage = NewSafeStorage()
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
	r.Get("/api/list", getLists)
	r.Get("/delete/{id}", deleteID)
	r.Post("/", WithLogging(func(w http.ResponseWriter, r *http.Request) {
		shortenedURL(w, r, cfg.BaseURL)
	}))
	r.Get("/{id}", WithLogging(redirectedURL))
	r.Post("/api/shorten", jsonHandler)

	fmt.Println(cfg.ServerAddress)
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		panic(err)
	}
}

// Функция для загрузки данных из файла
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
		storage.Set(rec.ShortURL, rec.OriginalURL)
		// Обновляем nextUUID, если найденное значение UUID больше текущего
		if id, err := strconv.Atoi(rec.UUID); err == nil && id >= nextUUID {
			nextUUID = id + 1
		}
	}
	return scanner.Err()
}

// Функция для добавления записи в файл
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

func shortenedURL(w http.ResponseWriter, r *http.Request, baseURL string) {
	// Проверка метода запроса
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
	storage.Set(shortURL, url)

	// Формирование записи с уникальным идентификатором
	record := Record{
		UUID:        strconv.Itoa(nextUUID),
		ShortURL:    shortURL,
		OriginalURL: url,
	}
	nextUUID++
	// Дописываем запись в файл
	if err := appendRecord(record, storageFilePath); err != nil {
		sugar.Errorf("Ошибка записи в файл: %v", err)
	}

	// Отправка ответа с сокращённым URL
	w.WriteHeader(201)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(baseURL + "/" + shortURL))
}

func redirectedURL(w http.ResponseWriter, r *http.Request) {
	// Проверка метода запроса
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
	// Поиск в карте
	value, exists := storage.Get(id)
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
		result += string(rune(rand.Intn(26) + 'a'))
	}
	return result
}

type (
	// структура для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// расширенный http.ResponseWriter для логирования
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

func jsonHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка метода запроса
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

	// Генерация короткого URL
	shortURL := genSym()
	storage.Set(shortURL, req.URL)

	record := Record{
		UUID:        strconv.Itoa(nextUUID),
		ShortURL:    shortURL,
		OriginalURL: req.URL,
	}
	nextUUID++
	if err := appendRecord(record, storageFilePath); err != nil {
		sugar.Errorf("Ошибка записи в файл: %v", err)
	}

	// Формирование ответа
	resp := redirectURL{
		Result: "http://localhost:8080/" + shortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Структуры для JSON
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

func getLists(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open(storageFilePath)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer file.Close()

	var records []Record

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var rec Record
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			sugar.Errorf("Ошибка парсинга записи: %v", err)
			continue
		}
		records = append(records, rec)
	}
	if scanner.Err() != nil {
		sugar.Errorf("Ошибка парсинга записи: %v", err)

	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(records); err != nil {
		http.Error(w, "Ошибка кодирования JSON", http.StatusInternalServerError)
	}
}

func deleteID(w http.ResponseWriter, r *http.Request) {
	// Получаем id из URL
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Некорректный id", http.StatusBadRequest)
		return
	}

	// Читаем все записи из файла
	file, err := os.Open(storageFilePath)
	if err != nil {
		http.Error(w, "Ошибка открытия файла", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var records []Record
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var rec Record
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			sugar.Errorf("Ошибка парсинга записи: %v", err)
			continue
		}
		records = append(records, rec)
	}
	if err := scanner.Err(); err != nil {
		sugar.Errorf("Ошибка чтения файла: %v", err)
		http.Error(w, "Ошибка чтения файла", http.StatusInternalServerError)
		return
	}

	// Фильтруем записи, исключая запись с заданным id
	var updatedRecords []Record
	var deleted bool
	for _, rec := range records {
		if rec.UUID == id {
			deleted = true
			// Также удаляем из in-memory хранилища
			storage.Delete(rec.ShortURL)
			continue
		}
		updatedRecords = append(updatedRecords, rec)
	}

	if !deleted {
		http.Error(w, "Запись не найдена", http.StatusNotFound)
		return
	}

	// Открываем файл для перезаписи (очищаем его)
	fileW, err := os.OpenFile(storageFilePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		http.Error(w, "Ошибка открытия файла для записи", http.StatusInternalServerError)
		return
	}
	defer fileW.Close()

	// Перезаписываем файл обновлёнными записями
	for _, rec := range updatedRecords {
		recBytes, err := json.Marshal(rec)
		if err != nil {
			sugar.Errorf("Ошибка маршалингу записи: %v", err)
			continue
		}
		if _, err := fileW.Write(append(recBytes, '\n')); err != nil {
			sugar.Errorf("Ошибка записи в файл: %v", err)
			http.Error(w, "Ошибка записи в файл", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Запись успешно удалена и файл перезаписан"))
}
