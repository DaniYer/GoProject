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

// --- Мок реализация интерфейса URLStoreWithDBforHandler ---

type MockStore struct {
	SaveFunc          func(shortURL, originalURL string) error
	GetFunc           func(shortURL string) (string, error)
	GetByOriginalFunc func(originalURL string) (string, error)
}

func (m *MockStore) Save(shortURL, originalURL string) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(shortURL, originalURL)
	}
	return nil
}

func (m *MockStore) Get(shortURL string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(shortURL)
	}
	return "", nil
}

func (m *MockStore) GetByOriginalURL(originalURL string) (string, error) {
	if m.GetByOriginalFunc != nil {
		return m.GetByOriginalFunc(originalURL)
	}
	return "", errors.New("not found")
}

// пустышка для полного интерфейса (иначе не соберётся):
func (m *MockStore) SaveWithConflict(shortURL, originalURL string) (string, error) {
	return "", nil
}

// --- Обязательно создаём мок-логгер для тестов ---

func setupLogger() {
	sugar = zap.NewNop().Sugar()
}

func TestGenerateShortURLHandler_Success(t *testing.T) {
	setupLogger()

	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	mockStore := &MockStore{}

	reqBody := "http://example.com"
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	GenerateShortURLHandler(rec, req, cfg, mockStore, sugar)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))

	respBody, _ := io.ReadAll(res.Body)
	assert.Contains(t, string(respBody), cfg.BaseURL+"/")
}

func TestGenerateShortURLHandler_Conflict(t *testing.T) {
	setupLogger()

	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	mockStore := &MockStore{
		GetByOriginalFunc: func(originalURL string) (string, error) {
			return "existingID", nil
		},
	}

	reqBody := "http://example.com"
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	GenerateShortURLHandler(rec, req, cfg, mockStore, sugar)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusConflict, res.StatusCode)

	respBody, _ := io.ReadAll(res.Body)
	assert.Equal(t, cfg.BaseURL+"/existingID", string(respBody))
}

func TestHandleShortenURLv13_Success(t *testing.T) {
	setupLogger()

	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	mockStore := &MockStore{}

	reqData := shortenRequest{URL: "http://example.com"}
	body, _ := json.Marshal(reqData)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	HandleShortenURLv13(rec, req, cfg, mockStore, sugar)

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
	setupLogger()

	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	mockStore := &MockStore{
		GetByOriginalFunc: func(originalURL string) (string, error) {
			return "existingID", nil
		},
	}

	reqData := shortenRequest{URL: "http://example.com"}
	body, _ := json.Marshal(reqData)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	HandleShortenURLv13(rec, req, cfg, mockStore, sugar)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusConflict, res.StatusCode)

	var resp shortenResponse
	err := json.NewDecoder(res.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, cfg.BaseURL+"/existingID", resp.Result)
}
