package shortener

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Мок-реализация интерфейса URLStoreWithDB для тестов
type MockStore struct {
	SaveWithConflictFunc func(shortURL, originalURL string) (string, error)
}

func (m *MockStore) Save(shortURL, originalURL string) error {
	return nil // не используется в этих тестах
}

func (m *MockStore) Get(shortURL string) (string, error) {
	return "", nil
}

func (m *MockStore) SaveWithConflict(shortURL, originalURL string) (string, error) {
	if m.SaveWithConflictFunc != nil {
		return m.SaveWithConflictFunc(shortURL, originalURL)
	}
	return shortURL, nil
}

func setupLogger() {
	sugar = zap.NewNop().Sugar()
}

func TestHandleShortenURL_Success(t *testing.T) {
	setupLogger()

	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	mockStore := &MockStore{
		SaveWithConflictFunc: func(shortURL, originalURL string) (string, error) {
			assert.NotEmpty(t, shortURL)
			assert.Equal(t, "http://example.com", originalURL)
			return shortURL, nil
		},
	}

	body := `{"url": "http://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	HandleShortenURL(rec, req, cfg, mockStore)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var resp redirectURL
	err := json.NewDecoder(res.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Contains(t, resp.Result, cfg.BaseURL+"/")
}

func TestHandleShortenURL_Conflict(t *testing.T) {
	setupLogger()

	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	mockStore := &MockStore{
		SaveWithConflictFunc: func(shortURL, originalURL string) (string, error) {
			return "existingShortURL", nil
		},
	}

	body := `{"url": "http://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	HandleShortenURL(rec, req, cfg, mockStore)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusConflict, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var resp redirectURL
	err := json.NewDecoder(res.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Contains(t, resp.Result, cfg.BaseURL+"/existingShortURL")
}

func TestHandleShortenURL_InvalidJSON(t *testing.T) {
	setupLogger()

	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	mockStore := &MockStore{}

	body := `invalid_json`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	HandleShortenURL(rec, req, cfg, mockStore)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandleShortenURL_SaveError(t *testing.T) {
	setupLogger()

	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	mockStore := &MockStore{
		SaveWithConflictFunc: func(shortURL, originalURL string) (string, error) {
			return "", errors.New("some db error")
		},
	}

	body := `{"url": "http://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	HandleShortenURL(rec, req, cfg, mockStore)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	responseBody, _ := io.ReadAll(res.Body)
	assert.Contains(t, string(responseBody), "Ошибка сохранения")
}
