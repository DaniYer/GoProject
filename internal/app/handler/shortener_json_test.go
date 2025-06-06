package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	"github.com/DaniYer/GoProject.git/internal/app/repository"
	"github.com/DaniYer/GoProject.git/internal/app/service"
	"github.com/stretchr/testify/assert"
)

// Создаем полностью in-memory хранилище для тестов
// Этот хэндлер будет использоваться для тестирования JSON API
func createTestHandlerJSON() *Handler {
	repo := repository.NewInMemoryRepository()
	svc := service.NewService(repo)

	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}

	// в тестах передаем заглушку pingFunc
	return NewHandler(svc, cfg, func(ctx context.Context) error { return nil })
}

func TestShortenJSONHandler(t *testing.T) {
	h := createTestHandlerJSON()

	body := []byte(`{"url": "https://practicum.yandex.ru"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ShortenJSONHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)

	var resp shortenResponse
	_ = json.NewDecoder(res.Body).Decode(&resp)

	assert.Contains(t, resp.Result, "http://localhost:8080/")
}
