package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
	"github.com/DaniYer/GoProject.git/internal/app/service"
	"github.com/stretchr/testify/assert"
)

type InMemoryMockStore struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewInMemoryMockStore() *InMemoryMockStore {
	return &InMemoryMockStore{
		data: make(map[string]string),
	}
}

func (m *InMemoryMockStore) Save(shortURL, originalURL string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[shortURL] = originalURL
	return shortURL, nil
}

func (m *InMemoryMockStore) GetByOriginalURL(originalURL string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for k, v := range m.data {
		if v == originalURL {
			return k, nil
		}
	}
	return "", errors.New("not found")
}

func (m *InMemoryMockStore) Get(shortURL string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	originalURL, ok := m.data[shortURL]
	if !ok {
		return "", errors.New("not found")
	}
	return originalURL, nil
}

func TestGenerateShortURLHandler_Success(t *testing.T) {
	store := NewInMemoryMockStore()
	svc := service.URLService{
		Store:   store,
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
	t.Logf("First create response: %s", string(respBody))
	assert.Contains(t, string(respBody), svc.BaseURL+"/")
}

func TestGenerateShortURLHandler_Conflict(t *testing.T) {
	store := NewInMemoryMockStore()
	svc := service.URLService{
		Store:   store,
		BaseURL: "http://localhost:8080",
	}

	reqBody := "http://example.com"

	req1 := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(reqBody))
	rec1 := httptest.NewRecorder()
	GenerateShortURLHandler(rec1, req1, &svc)
	res1 := rec1.Result()
	defer res1.Body.Close()
	io.ReadAll(res1.Body)

	req2 := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(reqBody))
	rec2 := httptest.NewRecorder()
	GenerateShortURLHandler(rec2, req2, &svc)
	res2 := rec2.Result()
	defer res2.Body.Close()

	assert.Equal(t, http.StatusConflict, res2.StatusCode)
}

func TestHandleShortenURLv13_Success(t *testing.T) {
	store := NewInMemoryMockStore()
	svc := service.URLService{
		Store:   store,
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
	store := NewInMemoryMockStore()
	svc := service.URLService{
		Store:   store,
		BaseURL: "http://localhost:8080",
	}

	reqData := dto.ShortenRequest{URL: "http://example.com"}
	body, _ := json.Marshal(reqData)

	req1 := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
	rec1 := httptest.NewRecorder()
	HandleShortenURLv13(rec1, req1, &svc)
	res1 := rec1.Result()
	defer res1.Body.Close()
	io.ReadAll(res1.Body)

	req2 := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
	rec2 := httptest.NewRecorder()
	HandleShortenURLv13(rec2, req2, &svc)
	res2 := rec2.Result()
	defer res2.Body.Close()

	assert.Equal(t, http.StatusConflict, res2.StatusCode)
}
