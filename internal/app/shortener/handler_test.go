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

// --- Общая функция логгера для тестов ---
func getTestLogger() *zap.SugaredLogger {
	return zap.NewNop().Sugar()
}

// --------------------------------------------------
// Мок для генератора (plain text / POST /)
// --------------------------------------------------

type MockStoreGenerator struct {
	SaveFunc          func(shortURL, originalURL string) error
	GetByOriginalFunc func(originalURL string) (string, error)
}

func (m *MockStoreGenerator) Save(shortURL, originalURL string) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(shortURL, originalURL)
	}
	return nil
}

func (m *MockStoreGenerator) GetByOriginalURL(originalURL string) (string, error) {
	if m.GetByOriginalFunc != nil {
		return m.GetByOriginalFunc(originalURL)
	}
	return "", errors.New("not found") // по умолчанию — не найдено
}

func (m *MockStoreGenerator) Get(shortURL string) (string, error) {
	return "", nil
}

// --------------------------------------------------
// Мок для JSON хендлера (POST /api/shorten)
// --------------------------------------------------

type MockStoreHandler struct {
	SaveFunc          func(shortURL, originalURL string) error
	GetByOriginalFunc func(originalURL string) (string, error)
}

func (m *MockStoreHandler) Save(shortURL, originalURL string) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(shortURL, originalURL)
	}
	return nil
}

func (m *MockStoreHandler) GetByOriginalURL(originalURL string) (string, error) {
	if m.GetByOriginalFunc != nil {
		return m.GetByOriginalFunc(originalURL)
	}
	return "", errors.New("not found")
}

func (m *MockStoreHandler) Get(shortURL string) (string, error) {
	return "", nil
}

// --------------------------------------------------
// ТЕСТЫ для генератора (plain text / POST /)
// --------------------------------------------------

func TestGenerateShortURLHandler_Success(t *testing.T) {
	logger := getTestLogger()
	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	mockStore := &MockStoreGenerator{}

	reqBody := "http://example.com"
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	GenerateShortURLHandler(rec, req, cfg, mockStore, logger)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))
	respBody, _ := io.ReadAll(res.Body)
	assert.Contains(t, string(respBody), cfg.BaseURL+"/")
}

func TestGenerateShortURLHandler_Conflict(t *testing.T) {
	logger := getTestLogger()
	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	mockStore := &MockStoreGenerator{
		GetByOriginalFunc: func(originalURL string) (string, error) {
			return "existingShortID", nil
		},
	}

	reqBody := "http://example.com"
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	GenerateShortURLHandler(rec, req, cfg, mockStore, logger)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusConflict, res.StatusCode)
	respBody, _ := io.ReadAll(res.Body)
	assert.Equal(t, cfg.BaseURL+"/existingShortID", string(respBody))
}

// --------------------------------------------------
// ТЕСТЫ для JSON (POST /api/shorten)
// --------------------------------------------------

func TestHandleShortenURLv13_Success(t *testing.T) {
	logger := getTestLogger()
	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	mockStore := &MockStoreHandler{}

	reqData := shortenRequest{URL: "http://example.com"}
	body, _ := json.Marshal(reqData)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	HandleShortenURLv13(rec, req, cfg, mockStore, logger)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var resp shortenResponse
	err := json.NewDecoder(res.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Contains(t, resp.Result, cfg.BaseURL+"/")
}

func TestHandleShortenURLv13_Conflict(t *testing.T) {
	logger := getTestLogger()
	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	mockStore := &MockStoreHandler{
		GetByOriginalFunc: func(originalURL string) (string, error) {
			return "existingShortID", nil
		},
	}

	reqData := shortenRequest{URL: "http://example.com"}
	body, _ := json.Marshal(reqData)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	HandleShortenURLv13(rec, req, cfg, mockStore, logger)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusConflict, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var resp shortenResponse
	err := json.NewDecoder(res.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, cfg.BaseURL+"/existingShortID", resp.Result)
}

func TestHandleShortenURLv13_InvalidJSON(t *testing.T) {
	logger := getTestLogger()
	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	mockStore := &MockStoreHandler{}

	body := []byte("invalid_json_data")

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	HandleShortenURLv13(rec, req, cfg, mockStore, logger)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}
