package shortener

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	"github.com/DaniYer/GoProject.git/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleShortenURL_Valid проверяет обработку корректного JSON-запроса,
// создание события и формирование правильного JSON-ответа.
func TestHandleShortenURL_Valid(t *testing.T) {
	cfg := &config.Config{B: "http://localhost:8080"}

	// Создаем временный файл для записи событий.
	tmpFile, err := os.CreateTemp("", "testproducer_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Создаем Producer с временным файлом.
	producer, err := storage.NewProducer(tmpFile.Name())
	require.NoError(t, err)

	// Формируем корректный JSON-запрос.
	reqBody := `{"url": "http://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	// Вызываем обработчик.
	HandleShortenURL(rec, req, cfg, producer)
	res := rec.Result()
	defer res.Body.Close()

	// Проверяем статус и заголовки.
	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	// Раскодируем JSON-ответ.
	var resp shortenResponse
	err = json.NewDecoder(res.Body).Decode(&resp)
	require.NoError(t, err)

	expectedPrefix := cfg.B + "/"
	assert.True(t, strings.HasPrefix(resp.Result, expectedPrefix), "result should start with %s", expectedPrefix)

	// Создаем Consumer для чтения событий из того же файла.
	consumer, err := storage.NewConsumer(tmpFile.Name())
	require.NoError(t, err)
	events, err := consumer.ReadEvents()
	require.NoError(t, err)
	require.NotEmpty(t, events)

	// Берем последнее записанное событие.
	lastEvent := events[len(events)-1]
	assert.Equal(t, "http://example.com", lastEvent.OriginalURL)

	shortIDFromResp := strings.TrimPrefix(resp.Result, expectedPrefix)
	assert.Equal(t, shortIDFromResp, lastEvent.ShortURL)
}

// TestHandleShortenURL_InvalidJSON проверяет, что при передаче некорректного JSON возвращается ошибка.
func TestHandleShortenURL_InvalidJSON(t *testing.T) {
	cfg := &config.Config{B: "http://localhost:8080"}

	tmpFile, err := os.CreateTemp("", "testproducer_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	producer, err := storage.NewProducer(tmpFile.Name())
	require.NoError(t, err)

	reqBody := "invalid json"
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	HandleShortenURL(rec, req, cfg, producer)
	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "Invalid JSON format")
}

// TestHandleShortenURL_WriteError проверяет, что при ошибке записи возвращается статус 500.
// Для симуляции ошибки мы закрываем файловый дескриптор через CloseFile().
func TestHandleShortenURL_WriteError(t *testing.T) {
	cfg := &config.Config{B: "http://localhost:8080"}

	tmpFile, err := os.CreateTemp("", "testproducer_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	producer, err := storage.NewProducer(tmpFile.Name())
	require.NoError(t, err)
	// Симулируем ошибку записи.
	require.NoError(t, producer.CloseFile())

	reqBody := `{"url": "http://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	HandleShortenURL(rec, req, cfg, producer)
	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "Failed to write event")
}

// TestGenerateShortURLHandler_Valid проверяет корректную работу обработчика, когда тело запроса – простой текст.
func TestGenerateShortURLHandler_Valid(t *testing.T) {
	cfg := &config.Config{B: "http://localhost:8080"}

	tmpFile, err := os.CreateTemp("", "testproducer_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	producer, err := storage.NewProducer(tmpFile.Name())
	require.NoError(t, err)

	reqBody := "http://example.org"
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	GenerateShortURLHandler(rec, req, cfg, producer)
	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	result := string(body)

	expectedPrefix := cfg.B + "/"
	assert.True(t, strings.HasPrefix(result, expectedPrefix), "result should start with %s", expectedPrefix)

	consumer, err := storage.NewConsumer(tmpFile.Name())
	require.NoError(t, err)
	events, err := consumer.ReadEvents()
	require.NoError(t, err)
	require.NotEmpty(t, events)

	lastEvent := events[len(events)-1]
	assert.Equal(t, "http://example.org", lastEvent.OriginalURL)

	shortIDFromResp := strings.TrimPrefix(result, expectedPrefix)
	assert.Equal(t, shortIDFromResp, lastEvent.ShortURL)
}

// TestGenerateShortURLHandler_WriteError проверяет, что при ошибке записи возвращается статус 500.
func TestGenerateShortURLHandler_WriteError(t *testing.T) {
	cfg := &config.Config{B: "http://localhost:8080"}

	tmpFile, err := os.CreateTemp("", "testproducer_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	producer, err := storage.NewProducer(tmpFile.Name())
	require.NoError(t, err)
	// Симулируем ошибку записи.
	require.NoError(t, producer.CloseFile())

	reqBody := "http://example.org"
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	GenerateShortURLHandler(rec, req, cfg, producer)
	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "Failed to write event")
}
