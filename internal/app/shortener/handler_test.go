package shortener

import (
	"bytes"
	"database/sql"
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

// создадим обёртку-адаптер для старого файлового стораджа под новый интерфейс Storage
type FileStorageAdapter struct {
	*storage.FileStorage
}

func (f *FileStorageAdapter) WriteEvent(event *storage.Event, db *sql.DB) error {
	return f.FileStorage.WriteEvent(event, nil)
}

func TestHandleShortenURL_Valid(t *testing.T) {
	cfg := &config.Config{B: "http://localhost:8080"}

	tmpFile, err := os.CreateTemp("", "test_filestorage_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fileStorage, err := storage.NewFileStorage(tmpFile.Name())
	require.NoError(t, err)
	adapter := &FileStorageAdapter{fileStorage}

	reqBody := `{"url": "http://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	HandleShortenURL(rec, req, cfg, adapter, nil)
	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var resp shortenResponse
	err = json.NewDecoder(res.Body).Decode(&resp)
	require.NoError(t, err)

	expectedPrefix := cfg.B + "/"
	assert.True(t, strings.HasPrefix(resp.Result, expectedPrefix))

	// Читаем события из файла
	consumer, err := storage.NewConsumer(tmpFile.Name())
	require.NoError(t, err)
	inMemory, err := consumer.ReadEvents()
	require.NoError(t, err)
	require.NotEmpty(t, inMemory.Data())

	found := false
	for _, event := range inMemory.Data() {
		if event.OriginalURL == "http://example.com" {
			found = true
			break
		}
	}
	assert.True(t, found, "original URL должен быть записан в файл")
}

func TestHandleShortenURL_InvalidJSON(t *testing.T) {
	cfg := &config.Config{B: "http://localhost:8080"}

	tmpFile, err := os.CreateTemp("", "test_filestorage_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fileStorage, err := storage.NewFileStorage(tmpFile.Name())
	require.NoError(t, err)
	adapter := &FileStorageAdapter{fileStorage}

	reqBody := "invalid json"
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	HandleShortenURL(rec, req, cfg, adapter, nil)
	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "Invalid JSON format")
}

func TestHandleShortenURL_WriteError(t *testing.T) {
	cfg := &config.Config{B: "http://localhost:8080"}

	tmpFile, err := os.CreateTemp("", "test_filestorage_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	fileStorage, err := storage.NewFileStorage(tmpFile.Name())
	require.NoError(t, err)
	require.NoError(t, fileStorage.CloseFile())

	adapter := &FileStorageAdapter{fileStorage}

	reqBody := `{"url": "http://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	HandleShortenURL(rec, req, cfg, adapter, nil)
	res := rec.Result()

	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "Failed to write event")
}
