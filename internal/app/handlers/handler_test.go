package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
	"github.com/DaniYer/GoProject.git/internal/app/middlewares"
	"github.com/DaniYer/GoProject.git/internal/app/service"
	"github.com/go-chi/chi/v5"
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

func (m *InMemoryMockStore) Save(shortURL, originalURL, userID string) (string, error) {
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

func (m *InMemoryMockStore) GetAllByUser(userID string) ([]dto.UserURL, error) {
	return nil, nil
}

func buildTestRouter(svc *service.URLService) http.Handler {
	r := chi.NewRouter()
	r.Use(middlewares.GzipHandle)
	r.Use(middlewares.InjectTestUserIDMiddleware("test-user-id"))
	r.Post("/api/shorten", NewHandleShortenURLv13(svc))
	return r
}

func TestHandleShortenURLv13_Success(t *testing.T) {
	store := NewInMemoryMockStore()
	svc := service.URLService{
		Store:   store,
		BaseURL: "http://localhost:8080",
	}

	router := buildTestRouter(&svc)

	reqData := dto.ShortenRequest{URL: "http://example.com"}
	body, _ := json.Marshal(reqData)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
	req.Header.Set("Accept-Encoding", "gzip") // ⚠ обязательно ставим gzip-заголовок
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var resp dto.ShortenResponse
	reader, _ := gzip.NewReader(res.Body)
	defer reader.Close()
	json.NewDecoder(reader).Decode(&resp)
	assert.Contains(t, resp.Result, svc.BaseURL+"/")
}
