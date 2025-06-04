package shortener

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
	"github.com/DaniYer/GoProject.git/internal/app/service"
	"github.com/stretchr/testify/assert"
)

// Универсальный мок на URLStore
type MockURLStore struct {
	SaveFunc          func(shortURL, originalURL string) (string, error)
	GetFunc           func(shortURL string) (string, error)
	GetByOriginalFunc func(originalURL string) (string, error)
}

func (m *MockURLStore) Save(shortURL, originalURL string) (string, error) {
	if m.SaveFunc != nil {
		return m.SaveFunc(shortURL, originalURL)
	}
	return shortURL, nil
}

func (m *MockURLStore) Get(shortURL string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(shortURL)
	}
	return "", nil
}

func (m *MockURLStore) GetByOriginalURL(originalURL string) (string, error) {
	if m.GetByOriginalFunc != nil {
		return m.GetByOriginalFunc(originalURL)
	}
	return "", errors.New("not found")
}

// Тесты для plain text (POST /)

func TestGenerateShortURLHandler_Success(t *testing.T) {
	mockStore := &MockURLStore{}
	svc := service.URLService{
		Store:   mockStore,
		BaseURL: "http://localhost:8080",
	}

	reqBody := "http://example.com"
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	GenerateShortURLHandler(rec, req, &svc)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))
	respBody, _ := io.ReadAll(res.Body)
	assert.Contains(t, string(respBody), svc.BaseURL+"/")
}

func TestGenerateShortURLHandler_Conflict(t *testing.T) {
	mockStore := &MockURLStore{
		GetByOriginalFunc: func(originalURL string) (string, error) {
			return "existingShortID", nil
		},
	}
	svc := service.URLService{
		Store:   mockStore,
		BaseURL: "http://localhost:8080",
	}

	reqBody := "http://example.com"
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	GenerateShortURLHandler(rec, req, &svc)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusConflict, res.StatusCode)
	respBody, _ := io.ReadAll(res.Body)
	assert.Equal(t, svc.BaseURL+"/existingShortID", string(respBody))
}

// Тесты для JSON (POST /api/shorten)

func TestHandleShortenURLv13_Success(t *testing.T) {
	mockStore := &MockURLStore{}
	svc := service.URLService{
		Store:   mockStore,
		BaseURL: "http://localhost:8080",
	}

	reqData := dto.ShortenRequest{URL: "http://example.com"}
	body, _ := json.Marshal(reqData)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	HandleShortenURLv13(rec, req, &svc)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var resp dto.ShortenResponse
	err := json.NewDecoder(res.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Contains(t, resp.Result, svc.BaseURL+"/")
}

func TestHandleShortenURLv13_Conflict(t *testing.T) {
	mockStore := &MockURLStore{
		GetByOriginalFunc: func(originalURL string) (string, error) {
			return "existingShortID", nil
		},
	}
	svc := service.URLService{
		Store:   mockStore,
		BaseURL: "http://localhost:8080",
	}

	reqData := dto.ShortenRequest{URL: "http://example.com"}
	body, _ := json.Marshal(reqData)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	HandleShortenURLv13(rec, req, &svc)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusConflict, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var resp dto.ShortenResponse
	err := json.NewDecoder(res.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, svc.BaseURL+"/existingShortID", resp.Result)
}

func TestHandleShortenURLv13_InvalidJSON(t *testing.T) {
	mockStore := &MockURLStore{}
	svc := service.URLService{
		Store:   mockStore,
		BaseURL: "http://localhost:8080",
	}

	body := []byte("invalid_json_data")

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	HandleShortenURLv13(rec, req, &svc)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}
